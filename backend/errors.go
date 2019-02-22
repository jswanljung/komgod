package backend

import (
	"errors"
)

var EmailAlreadyExistsError error  //nolint
var UserNameTakenError error       //nolint
var LoginFailedError error         // nolint
var UserNotVerifiedError error     // nolint
var NoSuchSessionError error       // nolint
var InvalidUserNameError error     //nolint
var InvalidEmailError error        //nolint
var PasswordLengthError error      //nolint
var InvalidScreenNameError error   //nolint
var UserAlreadyVerifiedError error //nolint

func init() {
	EmailAlreadyExistsError = errors.New("email already exists")
	UserNameTakenError = errors.New("username already taken")
	LoginFailedError = errors.New("login failed")
	UserNotVerifiedError = errors.New("account not verified")
	NoSuchSessionError = errors.New("no session matches token")
	InvalidUserNameError = errors.New("invalid format for username")
	InvalidEmailError = errors.New("invalid email address")
	PasswordLengthError = errors.New("password too long or short")
	InvalidScreenNameError = errors.New("invalid format for screenname")
	UserAlreadyVerifiedError = errors.New("user account already verified")
}
