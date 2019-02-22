package mejl

import (
	"bytes"
	"fmt"
	"log"
	"net/smtp"
	"text/template"
)

// Conf contains configuration data read by the main package
type Conf struct {
	Username string
	Password string
	SMTPhost string
	Port     int
	From     string
}

var (
	conf      *Conf
	vtemplate *template.Template
)

// Init loads initial template files. Needs
// config data, so must be called manually.
func Init(mconf *Conf, tpath string) {
	conf = mconf
	var err error
	vtemplate, err = template.ParseFiles(tpath + "/verifiera.mejl")
	if err != nil {
		log.Panic("Couldn't load verify mail template!")
	}
}

// SendVerificationMail sends an email with a verification link
func SendVerificationMail(recipient string, token string) (err error) {
	fmt.Println(fmt.Sprintf("%s:%d", conf.SMTPhost, conf.Port))
	auth := smtp.PlainAuth("", conf.Username, conf.Password,
		conf.SMTPhost)
	m := make(map[string]string)
	m["Recipient"] = recipient
	m["From"] = conf.From
	m["Link"] = "http://localhost:8080/kommentarer/verifieramejl/" + token
	var b bytes.Buffer
	err = vtemplate.Execute(&b, m)
	if err != nil {
		return
	}
	msg := b.Bytes()
	msg = bytes.Replace(msg, []byte{13, 10}, []byte{10}, -1)
	// replace CF \r (mac) with LF \n (unix)
	msg = bytes.Replace(msg, []byte{13}, []byte{10}, -1)
	msg = bytes.Replace(msg, []byte{10}, []byte{13, 10}, -1)
	to := []string{recipient}

	err = smtp.SendMail(fmt.Sprintf("%s:%d", conf.SMTPhost, conf.Port),
		auth, conf.From, to, msg)
	return
}
