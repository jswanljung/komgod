package backend

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

func clearTables() {
	_, err := db.Exec("DELETE FROM users;")
	if err != nil {
		panic(err)
	}
	db.Exec("DELETE FROM sessions;")
	db.Exec("DELETE FROM comments;")
}

// nolint
func TestAddUser(t *testing.T) {
	clearTables()
	name := "testus"
	email := "testus@grobjob.romcom"
	password := "The greatest fräcking pässwörd"
	screenName := "Nefarious Biter"
	token, _ := AddUser(name, email, password, screenName)
	if token == "" {
		t.Error("Null token")
	}
	// Is the user there?
	user, err := getUserFromEmail(email)
	if err != nil {
		t.Error(err)
	}

	err = confirmPassword(user.Pwdhash, password)
	if err != nil {
		t.Error(err)
	}

	if user.UserName != name || user.Email != email ||
		screenName != user.ScreenName {
		t.Error("User data incorrect")
	}

	// Try to add an identical user:
	tok, err := AddUser(name, email, password, screenName)
	if tok != "" {
		t.Errorf("Token for duplicate user: %s", tok)
	}

	if err != EmailAlreadyExistsError {
		t.Log(err)
		t.Error("Attempt to add duplicate user should give EmailAlreadyExistsError")
	}

	// Try to add a user with just the same user name

	_, err = AddUser(name, "bogus@boktog.romcom", "chrosh", "Tremendous Horticulturalist")
	if err != UserNameTakenError {
		t.Log(err)
		t.Error("Attempt to add with same username but not email should give UserNameTakenError")
	}

	// Try to add a user with just the same email:
	_, err = AddUser("SillyBilly", email, "snogdorff", "Rambunctious Perch")
	if err != EmailAlreadyExistsError {
		t.Log(err)
		t.Error("Attempt to add user with same email (but different name) should give EmailAlreadyExistsError")
	}
}

// Changed the login function to also return user. Test this too.
// nolint
func TestLogin(t *testing.T) {
	clearTables()
	name := "testus"
	email := "testus@grobjob.romcom"
	password := "The greatest fräcking pässwörd"
	screenName := "Nefarious Biter"
	token, _ := AddUser(name, email, password, screenName)
	VerifyAccount(token)
	name2 := "testina"
	email2 := "testina@jerkwork.sham"
	pwd2 := "No, mine's the greatest!"
	screenName2 := "Boffo much"
	AddUser(name2, email2, pwd2, screenName2)

	tok, _, err := Login(email, password, true)
	if tok == "" {
		t.Error("Null token")
	}
	if err != nil {
		t.Error(err)
	}
	tok, _, err = Login(name, password, true)
	if tok == "" {
		t.Error("Null token")
	}
	if err != nil {
		t.Error(err)
	}

	tok, _, err = Login(name2, pwd2, true)
	if tok != "" {
		t.Error("Token for unverified user!")
	}
	if err != UserNotVerifiedError {
		t.Log(err)
		t.Error("Unverified user login attempt should give UserNotVerifiedError")
	}

	tok, _, err = Login(name, pwd2, true)
	// Try using wrong password
	if tok != "" {
		t.Error("Token for wrong password")
	}
	if err != LoginFailedError {
		t.Log(err)
		t.Error("Wrong password should give LoginFailedError")

	}
}

func TestInsertComment(t *testing.T) {
	clearTables()
	name := "testus"
	email := "testus@grobjob.romcom"
	password := "The greatest fräcking pässwörd"
	screenName := "Nefarious Biter"
	token, _ := AddUser(name, email, password, screenName)
	VerifyAccount(token)
	path := "/komgo/"
	user, _ := getUserFromEmail(email)
	_ = InsertComment(0, path, "Dish ish a very schmart comment", user.ID)
	clist, _, _ := GetComments(path, user)
	id := clist[0].ID
	_ = InsertComment(id, path, "And dish ish a clever reply", user.ID)
	_ = InsertComment(id, path, "No it's not", user.ID)
	_ = InsertComment(0, path, "Screw all of you", user.ID)

	clist, last, _ := GetComments(path, user)
	t.Log(last)
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.Encode(clist)
	t.Log(buf.String())

}

//nolint
func TestValidation(t *testing.T) {
	ce := func(err error) {
		if err != nil {
			t.Error(err)
		}
	}
	cex := func(err error, msg string) {
		if err == nil {
			t.Error(msg)
		}
	}
	cex(validateUserName("df"), "should get error when username is too short")
	cex(validateUserName("sdfasdfasdasdfasdfas23dfasdfasdfasdfasd"), "should get error when uname too long")
	cex(validateUserName("asdfa."), "invalid chars in username should give error")
	ce(validateUserName("perfect12"))
	ce(validateUserName("Perfectly_sane"))
	ce(validateScreenName("Berta Ö'Brien"))
	cex(validateScreenName("Berta  O'brian"), "Multiple consecutive spaces not allowed")
	cex(validateScreenName("Berta O''Brian"), "Multiple consecutive apostrophes not ok.")
	ce(validateEmail("test@lisaforare.se"))
	cex(validateEmail("notanemail"), "Bad email doesn't give error")
	ce(validatePassword("This is a !\"Pärfektly gööd Pässword\""))
	cex(validatePassword("0"), "Too short password should give error")
	cex(validatePassword(`1234567890wekjföalsdkjf ölaksjdflaks askdlfj aöls askdfj öalskdjf alsdkjf öalsk
		asldfkjaösldkfja ölskdjfölak iweijalskdjfölkajsd fasldkfj aölskdfj öalksjdf ölakjsd flkads
		asldkfjaölsdjflaksjd fölkajsdölfkj alskdjfeiiqöwklfasdö flkja söldkfj aölskdjf laksjd fla
		alsdfjöalk9eia29ujnfpoi)asldkfjaölskdjfaölksdjflkasdlfjalksjdfwbekjnlasdjfljaslkdfjlkj`),
		"Alldeles för långt lösenord borde ge felmeddelande")
	cex(validateScreenName("a"), "för kort namn")
	cex(validateScreenName("Ett annars bra namn men lite för långt bara, eller vad vet jag?"), "För långt namn ska ge fel")
	ce(validateScreenName("Ett föredömligt namn1"))

	u, e, p, s, er := validateNewUserCredentials("barker1 ", "robot@b.com", "aösldieoo", "Im a barker")
	if u != "barker1" || e != "robot@b.com" || p != "aösldieoo" || s != "Im a barker" || er != nil {
		t.Errorf("Perfectly good credentials rejected: %s, %s, %s, %s, %s", u, e, p, s, er)
	}
	u, e, p, s, er = validateNewUserCredentials("barker1 p", "robot@b.com", "aösldieoo", "I'm a barker")
	if u != "" || e != "" || p != "" || s != "" || er == nil {
		t.Errorf("Bad credentials mishandled: %s", er)
	}
}

func TestMain(m *testing.M) {
	dbfile := "/Users/johan/test.db"
	Init(dbfile)
	out := m.Run()
	Close()
	//os.Remove(dbfile)
	os.Exit(out)

}
