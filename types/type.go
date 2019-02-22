package types

import (
	"log"
)

type role int

//nolint
const (
	BlockedRole    role = -1000
	UnverifiedRole role = -1
	DefaultRole    role = 0
	AdminRole      role = 1000
)

// User is a struct for user data
type User struct {
	UserName   string `json:"username"`
	Email      string `json:"-"`
	Role       role   `json:"role"`
	ScreenName string `json:"screenname"`
	Created    string `json:"-"`
	ID         int    `json:"-"`
	Pwdhash    string `json:"-"`
}

// IsVerified checks if the user is someone that has can comment
func (u User) IsVerified() bool {
	if u.Role == DefaultRole || u.Role == AdminRole {
		return true
	}
	return false
}

// IsUnverified checks specifically whether the user is not yet
// verified. Blocked users need not apply.
func (u User) IsUnverified() bool {
	return u.Role == UnverifiedRole
}

// IsAdmin checks if the user is an administrator
func (u User) IsAdmin() bool {
	if u.Role == AdminRole {
		return true
	}
	return false
}

// IsSomebody checks if the user is initialized
func (u User) IsSomebody() bool {
	return u.ID > 0
}

type status int

// nolint
const (
	VisibleStatus status = 1
	HiddenStatus  status = -1
)

// Comlist is a slice of Comment pointers with some methods
type Comlist []*Comment

// Comment is a type
type Comment struct {
	Path       string  `json:"-"`
	UserName   string  `json:"username"`
	ScreenName string  `json:"screenname"`
	Content    string  `json:"content"`
	Created    string  `json:"created"`
	ID         int     `json:"id"`
	Parent     int     `json:"parent"`
	Status     status  `json:"status"`
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
