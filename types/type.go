package types

import (
	"log"
)

// User is a struct for user data
type User struct {
	ID         int    `json:"-"`
	UserName   string `json:"username"`
	ScreenName string `json:"screenname"`
	Email      string `json:"-"`
	Pwdhash    string `json:"-"`
	Created    string `json:"-"`
	IsVerified bool   `json:"-"`
	IsAdmin    bool   `json:"isadmin"`
	IsBlocked  bool   `json:"isblocked"`
	WantsMail  bool   `json:"yestomail"`
}

// IsSomebody checks if the user is initialized
func (u User) IsSomebody() bool {
	return u.ID > 0
}

// Comlist is a slice of Comment pointers with some methods
type Comlist []*Comment

// Comment is a type
type Comment struct {
	ID         int     `json:"id"`
	Parent     int     `json:"parent"`
	Path       string  `json:"-"`
	Content    string  `json:"content"`
	UserName   string  `json:"username"`
	ScreenName string  `json:"screenname"`
	Created    string  `json:"created"`
	IsVisible  bool    `json:"isvisible"`
	Children   Comlist `json:"children"`
}

// AppendChild adds a child to a Comment
func (c *Comment) AppendChild(child *Comment) {
	c.Children = append(c.Children, child)
}

// AppendChildToLast adds a child ot the last comment in a Comlist
func (cl Comlist) AppendChildToLast(child *Comment) {
	//log.Print(cl)
	last := cl[len(cl)-1]
	if child.Parent == last.ID {
		cl[len(cl)-1].AppendChild(child)
	} else {
		log.Printf("Missing parent for comment %+v\n", *child)
	}
}
