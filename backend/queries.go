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
		email TEXT UNIQUE NOT NULL, 
		pwdhash TEXT NOT NULL, 
		role INTEGER DEFAULT (-1) NOT NULL,
		screenname TEXT NOT NULL,
		yestomail INTEGER DEFAULT (0) NOT NULL, 
		created DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL
		);`)

	commentsInitStmt = p(
		`CREATE TABLE IF NOT EXISTS comments (
		commentid INTEGER PRIMARY KEY AUTOINCREMENT UNIQUE, 
		parentid INTEGER, 
		path TEXT NOT NULL,
		content TEXT NOT NULL,
		userid INTEGER NOT NULL, 
		created DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL, 
		status INTEGER DEFAULT (1) NOT NULL,
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
		`SELECT username, email, role, screenname, created, userid, pwdhash
		FROM users WHERE email=?`)

	userFromUserNameStmt = p(
		`SELECT username, email, role, screenname, created, userid, pwdhash
		FROM users WHERE username=?`)

	addSessionStmt = p(
		`INSERT INTO sessions (
		token,
		userid,
		expire
		) VALUES (?, ?, ?);`)

	userFromTokenStmt = p(
		`SELECT 
		username,
		email,
		role,
		screenname,
		users.created,
		users.userid,
		pwdhash
	FROM users INNER JOIN sessions ON users.userid=sessions.userid WHERE token = ?;`)

	removeTokenStmt = p(
		`DELETE FROM sessions WHERE token=?;`)

	verifyUserStmt = p(
		`UPDATE users SET role=? WHERE userid=?;`)

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
		screenname,
		username,
		status,
		comments.created
		FROM comments INNER JOIN users ON comments.userid = users.userid WHERE 
		path=? AND comments.created > ? ORDER BY commentid;`)

	allCommentsStmt = p(
		`SELECT
		commentid,
		parentid,
		path,
		content,
		screenname,
		username,
		status,
		comments.created
		FROM comments INNER JOIN users ON comments.userid = users.userid WHERE 
		path=? ORDER BY rank;`)

	updateRankStmt = p(
		`UPDATE comments SET
		rank = CASE
		WHEN parentid = 0 THEN printf("%010u", commentid)
		ELSE printf("%010u-%010u", parentid, commentid)
		END
		WHERE rank IS NULL;`)

	return
}
