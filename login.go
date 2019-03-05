package main

import (

    //"fmt"

    "log"
    "net/http"
    "time"

    "github.com/jswanljung/komgod/backend"
    "github.com/jswanljung/komgod/types"
)

type LoginParams struct {
    Cred     string `json:"cred"`
    Password string `json:"password"`
    Remember bool   `json:"remember, omitempty"`
}

type LoginResponse struct {
    User types.User `json:"user"`
}

func login(lp LoginParams, w http.ResponseWriter) (resp *Response) {
    token, user, err := backend.Login(lp.Cred, lp.Password, lp.Remember)
    resp = new(Response)
    resp.Command = "login"
    if err != nil {
        switch err {
        case backend.UserNotVerifiedError:
            fallthrough
        case backend.LoginFailedError:
            errResp := new(HTMLResponse)
            errResp.applyTemplate("loginerror", nil)
            resp.Success = failed
            resp.Value = errResp
            return
        default:
            log.Panic(err)
        }
    }
    c := http.Cookie{Name: "login-token",
        Value:    token,
        Path:     "/",
        HttpOnly: true,
        SameSite: http.SameSiteStrictMode}
    if lp.Remember {
        c.Expires = time.Now().AddDate(0, 3, 0)
    }
    http.SetCookie(w, &c)
    lr := new(LoginResponse)
    lr.User = user
    resp.Success = succeeded
    resp.Value = lr
    return
}

//todo: forgotten password email

func removeLoginCookie(w http.ResponseWriter) {
    c := http.Cookie{Name: "login-token",
        Value:  "",
        MaxAge: -1}
    http.SetCookie(w, &c)
}

func logout(w http.ResponseWriter, token string) (resp *Response) {
    removeLoginCookie(w)
    if token != "" {
        err := backend.Logout(token)
        if err != nil {
            log.Panic(err)
        }
    }
    resp = new(Response)
    resp.Command = "logout"
    resp.Success = succeeded
    return
}
