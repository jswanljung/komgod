package main

import (

	//"fmt"

	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

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

	var rparams RequestParams
	dec := json.NewDecoder(r.Body)
	if dec.More() {
		err = dec.Decode(&rparams)
		if err != nil {
			log.Panic(err)
		}
	}
	/*err = r.Body.Close()
	if err != nil {
		log.Panic(err)
	}*/

	var resp *Response

	switch rparams.Command {
	case "newaccount":
		var rp NewAccountParams
		err = json.Unmarshal(rparams.Value, &rp)
		if err != nil {
			log.Panic(err)
		}
		resp = newAccount(rp)
	case "login":
		var lp LoginParams
		err = json.Unmarshal(rparams.Value, &lp)
		if err != nil {
			log.Panic(err)
		}
		resp = login(lp, w)
	case "logout":
		resp = logout(w, token)
	case "newcomment":
		var ncp NewCommentParams
		err = json.Unmarshal(rparams.Value, &ncp)
		if err != nil {
			log.Panic(err)
		}
		resp = newComment(ncp, r, user)
	default:
		resp = sendComments(r, 0, user)
	}

	resp.write(w)

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
