package backend

import (
	"database/sql"
	"log"
)

var (
	usersInitStmt,
	commentsInitStmt,
	sessionsInitStmt,
	addUserStmt,
	addSessionStmt,
	userFromEmailStmt,
	userFromUserNameStmt,
	removeTokenStmt,
	verifyUserStmt,
	insertCommentStmt,
	allCommentsStmt,
	commentsSinceStmt,
	updateRankStmt,
	blockUserStmt,
	unblockUserStmt,
	hideCommentStmt,
	unhideCommentStmt,
	userFromTokenStmt *sql.Stmt
)

func prepareStatements() (err error) {
	//defer recover()
	p := func(query string) (stmt *sql.Stmt) {
		stmt, err = db.Prepare(query)
		if err != nil {
			log.Fatalf("%s :query: %s", query, err)
		}
		return stmt
	}
	usersInitStmt = p(
		`CREATE TABLE IF NOT EXISTS users (
		userid INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE, 
		username TEXT UNIQUE NOT NULL, 
		screenname TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL, 
		pwdhash TEXT NOT NULL, 
		created DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
		isverified BOOL DEFAULT (FALSE) NOT NULL,
		isadmin BOOL DEFAULT (FALSE) NOT NULL,
		isblocked BOOL DEFAULT (FALSE) NOT NULL,
		wantsmail BOOL DEFAULT (FALSE) NOT NULL
		);`)

	commentsInitStmt = p(
		`CREATE TABLE IF NOT EXISTS comments (
		commentid INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE,
		parentid INTEGER, 
		path TEXT NOT NULL,
		content TEXT NOT NULL,
		userid INTEGER NOT NULL, 
		created DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL, 
		isvisible BOOL DEFAULT (TRUE) NOT NULL,
		rank TEXT
		);`)

	sessionsInitStmt = p(
		`CREATE TABLE IF NOT EXISTS sessions (
		token TEXT PRIMARY KEY NOT NULL UNIQUE,
		userid INTEGER NOT NULL,
		expire DATETIME NOT NULL,
		created DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`)

	_, err = usersInitStmt.Exec()
	if err != nil {
		log.Fatal("Failed to initialize users table")
	}
	_, err = sessionsInitStmt.Exec()
	if err != nil {
		log.Fatal("Failed to initialize sessions table")
	}
	_, err = commentsInitStmt.Exec()
	if err != nil {
		log.Fatal("Failed to initialize comments table")
	}

	addUserStmt = p(
		`INSERT INTO users (
		username,
		email,
		pwdhash,
		screenname
		) VALUES (?,?,?,?);
		`)

	userFromEmailStmt = p(
		`SELECT 
			userid,
			username, 
			screenname,
			email,
			pwdhash, 
			created,
			isverified,
			isadmin,
			wantsmail  
		FROM users WHERE email=?`)

	userFromUserNameStmt = p(
		`SELECT 
			userid,
			username, 
			screenname,
			email,
			pwdhash, 
			created,
			isverified,
			isadmin,
			wantsmail 
		FROM users WHERE username=?`)

	addSessionStmt = p(
		`INSERT INTO sessions (
		token,
		userid,
		expire
		) VALUES (?, ?, ?);`)

	userFromTokenStmt = p(
		`SELECT 
			users.userid,
			username, 
			screenname,
			email,
			pwdhash, 
			users.created,
			isverified,
			isadmin,
			wantsmail 
	FROM users INNER JOIN sessions ON users.userid=sessions.userid WHERE token = ?;`)

	removeTokenStmt = p(
		`DELETE FROM sessions WHERE token=?;`)

	verifyUserStmt = p(
		`UPDATE users SET isverified=TRUE WHERE userid=?;`)

	insertCommentStmt = p(
		`INSERT INTO comments (
			parentid,
			path,
			content,
			userid
			) VALUES (?,?,?,?);`)

	commentsSinceStmt = p(
		`SELECT
		commentid,
		parentid,
		path,
		content,
		username,
		screenname,
		comments.created,
		isvisible
		FROM comments INNER JOIN users ON comments.userid = users.userid WHERE 
		path=? AND commentid > ? ORDER BY commentid;`)

	allCommentsStmt = p(
		`SELECT
		commentid,
		parentid,
		path,
		content,
		username,
		screenname,
		comments.created,
		isvisible
		FROM comments INNER JOIN users ON comments.userid = users.userid WHERE 
		path=? ORDER BY rank;`)

	updateRankStmt = p(
		`UPDATE comments SET
		rank = CASE
		WHEN parentid = 0 THEN printf("%010u", commentid)
		ELSE printf("%010u-%010u", parentid, commentid)
		END
		WHERE rank IS NULL;`)

	blockUserStmt = p(
		`UPDATE users SET
		isblocked = TRUE WHERE userid = ?;`)

	unblockUserStmt = p(
		`UPDATE users SET
		isblocked = FALSE WHERE userid = ?;`)

	hideCommentStmt = p(
		`UPDATE comments SET
		isvisible = FALSE WHERE commentid = ?;`)

	unhideCommentStmt = p(
		`UPDATE comments SET
		isvisible = TRUE WHERE commentid = ?;`)

	return
}
