package main

import (

    //"fmt"

    "log"
    "net/http"
    "strings"

    "github.com/jswanljung/komgod/backend"
    "github.com/jswanljung/komgod/types"
)

func commentPath(r *http.Request) (path string) {
    return strings.TrimPrefix(r.URL.Path, "/kommentarer")
}

type CommentResponse struct {
    Comments types.Comlist `json:"comments, omitempty"`
    User     *types.User   `json:"user, omitempty"`
    LastID   int           `json:"lastid, omitempty"`
}

// TODO: rewrite to take only user and return a comment list
func sendComments(r *http.Request, lastID int, user types.User) (resp *Response) {
    resp = new(Response)
    resp.Command = "init"
    resp.Success = succeeded
    resp.Value = getCommentResponse(r, lastID, user)
    return
}

func getCommentResponse(r *http.Request, lastID int, user types.User) (cr *CommentResponse) {
    comments, last, err := backend.GetCommentsSince(commentPath(r), lastID, user)
    if err != nil {
        log.Panic(err)
    }
    cr = new(CommentResponse)
    if user.IsSomebody() && user.IsVerified {
        cr.User = &user
    }
    cr.Comments = comments
    cr.LastID = last
    return
}

type NewCommentParams struct {
    Content string `json:"content"`
    Parent  int    `json:"parent"`
    LastID  int    `json:"lastid"`
}

// Todo split into newComment and commentsSince. Implement some kind of error response?
func newComment(ncp NewCommentParams, r *http.Request, user types.User) (resp *Response) {
    if strings.TrimSpace(ncp.Content) == "" {
        log.Panic("Empty comment not permitted!")
    }
    err := backend.InsertComment(ncp.Parent, commentPath(r), ncp.Content, user.ID)
    //todo: proper error handling
    if err != nil {
        log.Panic(err)
    }
    resp = new(Response)
    resp.Command = "newcomment"
    resp.Success = succeeded
    resp.Value = getCommentResponse(r, ncp.LastID, user)
    return

}
