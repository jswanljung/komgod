package main

import (

    //"fmt"

    "bytes"
    "log"

    "github.com/jswanljung/komgod/backend"
    "github.com/jswanljung/komgod/mejl"
)

type NewAccountParams struct {
    UserName   string `json:"username"`
    Email      string `json:"email"`
    ScreenName string `json:"screenname"`
    Password   string `json:"password"`
    WantsEmail bool   `json:"wantsemail, omitempty"`
}

type HTMLResponse struct {
    HTML string `json:"html"`
}

func (h *HTMLResponse) applyTemplate(tname string, data interface{}) {
    var buf bytes.Buffer
    err := mallar.ExecuteTemplate(&buf, tname, data)
    if err != nil {
        log.Panic(err)
    }
    h.HTML = buf.String()
}

func newAccount(p NewAccountParams) (r *Response) {
    r = new(Response)
    r.Command = "newaccount"
    h := new(HTMLResponse)
    token, err := backend.AddUser(p.UserName, p.Email, p.Password, p.ScreenName)
    var template string
    switch err {
    case nil:
        er := mejl.SendVerificationMail(p.Email, token)
        if er != nil {
            template = "generalerror"
        } else {
            r.Success = succeeded
            template = "kontoskapat"
        }

    case backend.EmailAlreadyExistsError:
        template = "emailalreadyexists"
    case backend.UserNameTakenError:
        template = "useralreadyexists"
    // TODO: handle validation errors separately?
    default:
        template = "generalerror"
    }
    h.applyTemplate(template, p)
    r.Value = h
    return
}
