package main

import (
	"bytes"
	"encoding/json"
	//"fmt"
	"github.com/BurntSushi/toml"
	"github.com/jswanljung/komgo/database"
	"github.com/jswanljung/komgo/mejl"
	"github.com/jswanljung/komgo/types"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

//todo: read config file with templates
//todo: use templates to spit something out.

// Define a struct for configuration

type Config struct {
	BaseTemplate string
	TemplateDir  string
	MailConf     mejl.MejlConf
	DbFile       string
}

var conf Config
var errTemplate *template.Template
var mallar *template.Template

func getNewAccountHandler() func(http.ResponseWriter, *http.Request) {
	t, err := template.ParseFiles(conf.BaseTemplate, conf.TemplateDir+"/nyttkonto.html")
	if err != nil {
		log.Panic("Error loading new account template!")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		t.Execute(w, nil)
	}
}

func getAccountCreatedHandler() func(http.ResponseWriter, *http.Request) {
	t, err := template.ParseFiles(conf.BaseTemplate, conf.TemplateDir+"/skapakonto.html")
	if err != nil {
		log.Panic("Error loading create account template!")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		//todo: skapa konto
		// finns det redan ett konto med den mejladressen?
		// är det en någorlunda vettig mejladress med @?
		// skicka då mejl.
		var m map[string]string
		m = make(map[string]string)
		m["Email"] = r.PostFormValue("email")
		mejl.SendVerificationMail(m["Email"], "abcde")
		err = t.Execute(w, m)
		if err != nil {
			log.Panic(err)
		}
	}
}

func getVerificationHandler() func(http.ResponseWriter, *http.Request) {
	t, err := template.ParseFiles(conf.BaseTemplate, conf.TemplateDir+"/verifierad.html")
	if err != nil {
		log.Panic("Error loading verified template!")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// Användaren är kanske redan verifierad?
		// eller så känner vi inte igen verifikationskoden
		var m map[string]string
		token := strings.TrimPrefix(r.URL.Path, "/kommentarer/verifieramejl/")
		err = database.VerifyUser(token)
		if err != nil {
			panic(err)
		}
		err = t.Execute(w, m)
		if err != nil {
			log.Panic(err)
		}
	}
}

type response struct {
	Header string `json:"header"`
	HTML   string `json:"html"`
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	encoder := json.NewEncoder(w)
	encoder.Encode(data)
}

func writeJSONResponse(w http.ResponseWriter, header, templatename string, data interface{}) {
	var newr response
	newr.Header = header
	var buf bytes.Buffer
	if err := mallar.ExecuteTemplate(&buf, templatename, data); err != nil {
		panic(err)
	}
	newr.HTML = buf.String()
	writeJSON(w, newr)
}

func handler(w http.ResponseWriter, r *http.Request) {
	//Do authentication here, also do a recover
	defer func() {
		if rec := recover(); rec != nil {
			writeJSONResponse(w, "Nåt blev fel", "generalerror", nil)
		}
	}()
	var token string
	var user types.User
	cookie, err := r.Cookie("login-token")
	if err == nil {
		token = cookie.Value
		user, err = database.GetUserFromToken(token)
		if err != nil {
			removeLoginCookie(w)
		}
	}

	log.Print(r.PostFormValue("action"))
	switch r.PostFormValue("action") {

	case "newaccount":
		newAccount(w, r)
	case "login":
		login(w, r)
	case "logout":
		logout(w, token)
	default:
		m := make(map[string]interface{})
		if user.UserName != "" {
			m["user"] = user
		} else {
			m["status"] = "0"
		}
		writeJSON(w, m)
		// Here we print comments.
	}

}

func newAccount(w http.ResponseWriter, r *http.Request) {
	var m map[string]string
	defer func() {
		if rec := recover(); rec != nil {
			if er, ok := rec.(database.Error); ok {
				switch er.Code {
				case database.EmailAlreadyExistsError:
					writeJSONResponse(w, "Kontot finns redan!", "emailalreadyexists", m)
				case database.UserNameTakenError:
					writeJSONResponse(w, "Användarnamnet upptaget!", "useralreadyexists", m)
				default:
					writeJSONResponse(w, "Nåt blev fel!", "generalerror", m)
				}
			} else {
				panic(rec)
			}
		}
	}()
	m = map[string]string{"Email": r.FormValue("email")}
	m["Username"] = r.FormValue("username")
	m["Password"] = r.FormValue("password")
	m["Screenname"] = r.FormValue("screenname")
	token, err := database.AddUser(m["Username"], m["Email"], m["Password"], m["Screenname"])
	if err != nil {
		panic(err)
	}
	mejl.SendVerificationMail(m["Email"], token)
	writeJSONResponse(w, "Kolla din inbox!", "kontoskapat", m)
}

func login(w http.ResponseWriter, r *http.Request) {
	cred := r.PostFormValue("cred")
	password := r.PostFormValue("password")
	remember := r.PostFormValue("remember")
	token, err := database.Login(cred, password, remember == "on")
	if err != nil {
		panic(err)
	}
	c := http.Cookie{Name: "login-token",
		Value:    token,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode}
	if remember == "on" {
		c.Expires = time.Now().AddDate(0, 3, 0)
	}
	http.SetCookie(w, &c)
	user, err := database.GetUserFromToken(token)
	if err != nil {
		panic(err)
	}
	m := map[string]interface{}{"user": user}
	writeJSON(w, m)
}

//todo: forgotten password email

//todo: logout

func removeLoginCookie(w http.ResponseWriter) {
	c := http.Cookie{Name: "login-token",
		Value:  "",
		MaxAge: -1}
	http.SetCookie(w, &c)
}

func logout(w http.ResponseWriter, token string) {
	removeLoginCookie(w)
	if token != "" {
		database.Logout(token)
	}
	m := map[string]string{"status": "ok"}
	writeJSON(w, m)
}

func main() {
	//os.Setenv("KOMMENTCONFIG", "/Users/johan/komment.config")
	conffile, confexists := os.LookupEnv("KOMMENTCONFIG")
	if !confexists {
		conffile = os.ExpandEnv("$HOME/komment.config")
	}
	// I guess we could do some error handling here too.
	toml.DecodeFile(conffile, &conf)
	mallar = template.Must(template.ParseFiles(conf.TemplateDir + "/mallar.html"))
	mejl.Init(&(conf.MailConf), conf.TemplateDir)
	database.Init(conf.DbFile)
	/*var err error
	errTemplate, err = template.ParseFiles(conf.BaseTemplate, conf.TemplateDir+"/error.html")
	if err != nil {
		log.Panic("Error loading error template!")
	}
	http.HandleFunc("/kommentarer/nyttkonto", getNewAccountHandler())
	http.HandleFunc("/kommentarer/skapakonto", getAccountCreatedHandler())
	*/
	http.HandleFunc("/kommentarer/verifieramejl/", getVerificationHandler())
	http.HandleFunc("/kommentarer/", handler)
	log.Fatal(http.ListenAndServe(":12000", nil))
}
