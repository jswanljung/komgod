package main

import (

	//"fmt"
	"bytes"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/jswanljung/komgod/backend"
	"github.com/jswanljung/komgod/mejl"
	"github.com/jswanljung/komgod/types"
)

//todo: get rid of hardcoded paths

// Define a struct for configuration

type config struct {
	BaseTemplate string
	TemplateDir  string
	MailConf     mejl.Conf
	DbFile       string
}

var conf config
var mallar *template.Template

var verifyTemplate *template.Template
var errorTemplate *template.Template

func dontPanic(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Print(r)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	defer dontPanic(w)
	var m map[string]string
	token := strings.TrimPrefix(r.URL.Path, "/kommentarer/verifieramejl/")
	err := backend.VerifyAccount(token)
	if err != nil {
		switch err {
		case backend.NoSuchSessionError:
			m = map[string]string{"Title": "Kunde inte verifiera mejladressen!",
				"Message": "Verifieringskoden är bara giltig i 30 dagar. Prova att skicka ett nytt."}

		case backend.UserAlreadyVerifiedError:
			m = map[string]string{"Title": "Adressen redan verifierad!",
				"Message": "Du har redan verifierat den här mejladressen. Prova att logga in."}
		default:
			log.Panic(err)
		}
		err = errorTemplate.Execute(w, m)
		if err != nil {
			log.Panic(err)
		}
	}
	err = verifyTemplate.Execute(w, m)
	if err != nil {
		log.Panic(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	defer dontPanic(w)
	var token string
	var user types.User
	cookie, err := r.Cookie("login-token")
	if err == nil {
		token = cookie.Value
		user, err = backend.GetUserFromToken(token)
		if err != nil {
			removeLoginCookie(w)
		}
	}

	switch r.PostFormValue("action") {
	case "newaccount":
		newAccount(w, r)
	case "login":
		login(w, r)
	case "logout":
		logout(w, token)
	case "newcomment":
		newComment(w, r, user)
	default:
		sendComments(w, r, user)
	}

}

func sendComments(w http.ResponseWriter, r *http.Request, user types.User) {

	now := time.Now().UTC()
	comments, err := backend.GetComments(commentPath(r), user)
	if err != nil {
		log.Panic(err)
	}
	jr := newJSONResponse()
	if user.IsSomebody() && user.IsVerified() {
		jr.addUser(user)
	}
	jr.add("before", now).add("comments", comments).write(w)
}

func newComment(w http.ResponseWriter, r *http.Request, user types.User) {
	content := r.FormValue("content")
	parent := r.FormValue("parent")
	since, err := time.Parse(time.RFC3339, r.FormValue("since"))
	if err != nil {
		log.Panic(err)
	}
	var pid int64
	if parent != "" {
		pid, err = strconv.ParseInt(parent, 10, 64)
	}
	if err != nil {
		log.Panic(err)
	}
	err = backend.InsertComment(int(pid), commentPath(r), content, user.ID)
	//todo: proper error handling
	if err != nil {
		log.Panic(err)
	}
	comments, err := backend.GetCommentsSince(commentPath(r), since, user)
	now := time.Now().UTC()
	if err != nil {
		log.Panic(err)
	}
	newJSONResponse().addUser(user).add("before", now).add("comments", comments).write(w)
}

func newAccount(w http.ResponseWriter, r *http.Request) {
	m := map[string]string{"Email": r.FormValue("email")}
	m["Username"] = r.FormValue("username")
	m["Password"] = r.FormValue("password")
	m["Screenname"] = r.FormValue("screenname")
	token, err := backend.AddUser(m["Username"], m["Email"], m["Password"], m["Screenname"])
	if err != nil {
		switch err {
		case backend.EmailAlreadyExistsError:
			newJSONResponse().addTemp("emailalreadyexists", m).write(w)
		case backend.UserNameTakenError:
			newJSONResponse().addTemp("useralreadyexists", m).write(w)
		// TODO: handle validation errors separately?
		default:
			newJSONResponse().addTemp("generalerror", m).write(w)
		}
		return
	}
	err = mejl.SendVerificationMail(m["Email"], token)
	if err != nil {
		newJSONResponse().addTemp("generalerror", m).write(w)
		return
	}
	newJSONResponse().addTemp("kontoskapat", m).write(w)
}

func login(w http.ResponseWriter, r *http.Request) {
	cred := r.PostFormValue("cred")
	password := r.PostFormValue("password")
	remember := r.PostFormValue("remember")
	token, user, err := backend.Login(cred, password, remember == "on")
	if err != nil {
		switch err {
		case backend.UserNotVerifiedError:

		case backend.LoginFailedError:
			newJSONResponse().addTemp("loginerror", nil).write(w)
		default:
			log.Panic(err)
		}
	}
	c := http.Cookie{Name: "login-token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode}
	if remember == "on" {
		c.Expires = time.Now().AddDate(0, 3, 0)
	}
	http.SetCookie(w, &c)
	newJSONResponse().addUser(user).write(w)
}

//todo: forgotten password email

func removeLoginCookie(w http.ResponseWriter) {
	c := http.Cookie{Name: "login-token",
		Value:  "",
		MaxAge: -1}
	http.SetCookie(w, &c)
}

func logout(w http.ResponseWriter, token string) {
	removeLoginCookie(w)
	if token != "" {
		err := backend.Logout(token)
		if err != nil {
			log.Panic(err)
		}
	}
	newJSONResponse().statusOK().write(w)
}

func commentPath(r *http.Request) (path string) {
	return strings.TrimPrefix(r.URL.Path, "/kommentarer")
}

func main() {
	//os.Setenv("KOMMENTCONFIG", "/Users/johan/komment.config")
	conffile, confexists := os.LookupEnv("KOMMENTCONFIG")
	if !confexists {
		conffile = os.ExpandEnv("$HOME/komment.config")
	}
	// I guess we could do some error handling here too.
	_, err := toml.DecodeFile(conffile, &conf)
	if err != nil {
		log.Fatal("Could not read config file!")
	}
	mallar = template.Must(template.ParseFiles(conf.TemplateDir + "/mallar.html"))
	mejl.Init(&(conf.MailConf), conf.TemplateDir)
	backend.Init(conf.DbFile)
	defer func() {
		backend.Close()
	}()

	errorTemplate, err = template.ParseFiles(conf.BaseTemplate, conf.TemplateDir+"/error.html")
	if err != nil {
		log.Panic("Error loading error template!")
	}
	verifyTemplate, err = template.ParseFiles(conf.BaseTemplate, conf.TemplateDir+"/verifierad.html")
	if err != nil {
		log.Panic("Error loading verify template!")
	}

	http.HandleFunc("/kommentarer/verifieramejl/", verifyHandler)
	http.HandleFunc("/kommentarer/", handler)
	log.Fatal(http.ListenAndServe(":12000", nil))
}

// Helper type for sending responses. Could probably
// be made smarter
type jSONResponse map[string]interface{}

func newJSONResponse() jSONResponse {
	return make(jSONResponse)
}

func (j jSONResponse) write(w http.ResponseWriter) {
	encoder := json.NewEncoder(w)
	err := encoder.Encode(j)
	if err != nil {
		log.Panic(err)
	}
}

func (j jSONResponse) addTemp(tname string, data interface{}) jSONResponse {
	var buf bytes.Buffer
	err := mallar.ExecuteTemplate(&buf, tname, data)
	if err != nil {
		log.Panic(err)
	}
	j["html"] = buf.String()
	return j
}

func (j jSONResponse) add(key string, value interface{}) jSONResponse {
	j[key] = value
	return j
}

func (j jSONResponse) addUser(user interface{}) jSONResponse {
	j["user"] = user
	return j
}

func (j jSONResponse) statusOK() jSONResponse {
	j["status"] = "ok"
	return j
}
