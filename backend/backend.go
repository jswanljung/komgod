package backend

import (
	"database/sql"
	"log"
	"time"

	"github.com/jswanljung/komgod/types"
	_ "github.com/mattn/go-sqlite3" //nolint
)

var db *sql.DB

// Init just opens the database file and prepares prepared statements. Failure is fatal.
// This has to be called manually since it needs a filename
func Init(dbfile string) {
	var err error
	db, err = sql.Open("sqlite3", dbfile)
	if err != nil {
		log.Fatal(err)
	}
	err = prepareStatements()
	if err != nil {
		log.Fatal(err)
	}
}

// Close closes the database file.
func Close() {
	err := db.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func scanUser(row *sql.Row, user *types.User) (err error) {
	err = row.Scan(
		&user.ID,
		&user.UserName,
		&user.ScreenName,
		&user.Email,
		&user.Pwdhash,
		&user.Created,
		&user.IsVerified,
		&user.IsAdmin,
		&user.WantsMail)
	return
}

func getUserFromEmail(email string) (user types.User, err error) {
	row := userFromEmailStmt.QueryRow(email)
	err = scanUser(row, &user)
	return
}

func getUserFromUserName(username string) (user types.User, err error) {
	row := userFromUserNameStmt.QueryRow(username)
	err = scanUser(row, &user)
	return
}

func getUserFromEmailOrUserName(emailorusername string) (user types.User, err error) {
	verr := validateEmail(emailorusername)
	if verr != nil {
		return getUserFromUserName(emailorusername)
	}
	return getUserFromEmail(emailorusername)
}

func newToken(user types.User, keep bool) (token string, err error) {
	token = sessionToken()
	expiry := time.Now().AddDate(0, 0, 1)
	if keep {
		expiry = expiry.AddDate(0, 1, 0)
	}
	_, err = addSessionStmt.Exec(token, user.ID, expiry.UTC())
	return
}

// GetUserFromToken returns user data for a session token
// Errors: NoSuchSessionError, evil database error
func GetUserFromToken(token string) (user types.User, err error) {
	row := userFromTokenStmt.QueryRow(token)
	err = scanUser(row, &user)
	if err == sql.ErrNoRows {
		err = NoSuchSessionError
	}
	return
}

// Login checks a user login and returns a session token on success
// Possible errors: LoginFailedError, UserNotVerifiedError or a database error
func Login(emailorusername string, password string, remember bool) (token string,
	user types.User, err error) {
	user, err = getUserFromEmailOrUserName(emailorusername)
	if err != nil {
		if err == sql.ErrNoRows {
			err = LoginFailedError
		} else {
			return
		}
	}
	// returns LoginFailedError if password doesn't match
	err = confirmPassword(user.Pwdhash, password)
	if err != nil {
		return
	}
	if !user.IsVerified {
		err = UserNotVerifiedError
		return
	}
	token, err = newToken(user, remember)
	return
}

// VerifyAccount checks to see if session token matches and if so
// upgrades user status to verified.
// Errors: NoSuchSessionError, UserAlreadyVerifiedError, evil database error
func VerifyAccount(token string) (err error) {
	user, err := GetUserFromToken(token)
	if err != nil {
		return
	}
	if user.IsVerified {
		err = UserAlreadyVerifiedError
		return
	}
	_, err = verifyUserStmt.Exec(user.ID)
	return
}

// Logout deletes the user session
// Error: could return a database error
func Logout(token string) (err error) {
	_, err = removeTokenStmt.Exec(token)
	return
}

// AddUser runs some validation, then inserts a new user and creates a new session
// (for email validation), returning the session token
func AddUser(username, email, password, screenname string) (token string, err error) {
	// First, perform some validation (also strips whitespace)
	u, e, p, s, err := validateNewUserCredentials(username,
		email, password, screenname)
	if err != nil {
		return // a validation error
	}
	pwdhash := passwordHash(p)
	// Now we check to make sure this user doesn't exist already.
	if _, okerr := getUserFromEmail(e); okerr == nil {
		err = EmailAlreadyExistsError
	}
	if err != nil {
		return
	}
	if _, okerr := getUserFromUserName(u); okerr == nil {
		err = UserNameTakenError
		return
	}
	_, err = addUserStmt.Exec(u, e, pwdhash, s)
	if err != nil {
		return
	}
	user, err := getUserFromEmail(e)
	if err != nil {
		return
	}
	token, err = newToken(user, true)
	return
}

// GetComments returns a tree of all comments visible to user
func GetComments(path string, user types.User) (comments types.Comlist, lastID int, err error) {
	return GetCommentsSince(path, 0, user)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetCommentsSince returns a list of all comments visible to user since time. If
// since.IsZero(), returns a tree, otherwise just a list.
func GetCommentsSince(path string, sinceID int, user types.User) (comments types.Comlist, lastID int, err error) {
	var rows *sql.Rows
	asTree := false
	if sinceID == 0 {
		rows, err = allCommentsStmt.Query(path)
		asTree = true
	} else {
		rows, err = commentsSinceStmt.Query(path, sinceID)
	}
	if err != nil {
		return
	}

	for rows.Next() {
		var newComment types.Comment
		err = rows.Scan(
			&newComment.ID,
			&newComment.Parent,
			&newComment.Path,
			&newComment.Content,
			&newComment.UserName,
			&newComment.ScreenName,
			&newComment.Created,
			&newComment.IsVisible)
		if err != nil {
			continue
		}

		lastID = max(lastID, newComment.ID)
		if user.IsAdmin || newComment.IsVisible {
			if asTree && newComment.Parent != 0 {
				comments.AppendChildToLast(&newComment)
			} else {
				comments = append(comments, &newComment)
			}
		}
	}
	if err != nil {
		return
	}
	// An empty result is normal if there are no comments on the page
	// Other database errors are not okay.
	if e := rows.Err(); e != sql.ErrNoRows {
		err = e
	}
	return
}

//InsertComment puts a new comment in the database
func InsertComment(parent int, path string, content string, userid int) (err error) {
	_, err = insertCommentStmt.Exec(parent, path, content, userid)
	if err != nil {
		return
	}
	_, err = updateRankStmt.Exec()
	return
}
