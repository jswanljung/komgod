package backend

import (
	"regexp"
	"strings"

	"github.com/badoux/checkmail"
)

func validateNewUserCredentials(username,
	email, password, screenname string) (u, e, p, s string, err error) {
	username = strings.TrimSpace(username)
	err = validateUserName(username)
	if err != nil {
		return
	}
	email = strings.TrimSpace(email)
	err = validateEmail(email)
	if err != nil {
		return
	}
	password = strings.TrimSpace(password)
	err = validatePassword(password)
	if err != nil {
		return
	}
	screenname = strings.TrimSpace(screenname)
	err = validateScreenName(screenname)
	if err != nil {
		return
	}
	u = username
	e = email
	p = password
	s = screenname
	return
}

var uNameRegexp = regexp.MustCompile(`\A[\pL\pN_]{4,30}\z`)

func validateUserName(username string) (err error) {
	/* Only unicode letters and numbers. length between 4 and 30 chars */
	if len(username) < 4 || len(username) > 30 || !uNameRegexp.MatchString(username) {
		err = InvalidUserNameError
	}
	return
}

func validateEmail(email string) (err error) {
	err = checkmail.ValidateFormat(email)
	if err != nil {
		err = InvalidEmailError
	}
	return
}

func validatePassword(password string) (err error) {
	/* Min 5, max 200, otherwise who cares? */
	if len(password) < 5 || len(password) > 200 {
		err = PasswordLengthError
	}
	return
}

//^(?:[A-Za-z0-9\u00C0-\u00FF]+[' ]?)*$
var sNameRegexp = regexp.MustCompile(`\A(?:[\pL\pN]+['\pZs]?)+\z`)

func validateScreenName(screenname string) (err error) {
	/* Min 2, max 30, only unicode letters, numbers and spaces. */
	if len(screenname) < 2 || len(screenname) > 30 || !sNameRegexp.MatchString(screenname) {
		err = InvalidScreenNameError
	}
	return
}
