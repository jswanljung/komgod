package main

import (

    //"fmt"

    "log"
    "net/http"
    "strings"

    "github.com/jswanljung/komgod/backend"
)

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
        return
    }
    err = verifyTemplate.Execute(w, m)
    if err != nil {
        log.Panic(err)
    }
}
