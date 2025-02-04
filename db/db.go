package db

import (
    "gotard/auth"
	"database/sql"
	"log"
	"time"
	"context"
	_ "github.com/mattn/go-sqlite3"
)

type User struct {
	Email 		string
	UserName 	string
	CreatedAt 	int64
	Password 	string // hashed, of course
}

var DB *sql.DB

func InitDB(ctx context.Context, dataSourceName string) error {

	var err error
	DB, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
        return err
    }
    if err = DB.Ping(); err != nil {
        return err
    }

   err = _createUserTable(ctx)
   if err != nil {
       log.Fatal("Error creating user table:", err)
       return err
   }

   err = _createSessionTable(ctx)
   if err != nil {
       log.Fatal("Error creating session table:", err)
       return err
   }
   return nil
}

func _createUserTable(ctx context.Context) error {
	tableSQL := `CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
	    username TEXT NOT NULL,
	    created_at INTEGER NOT NULL,
	    password TEXT NOT NULL,
        sub_tier SMALLINT DEFAULT 0
	);`

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
        return err
    }
    defer tx.Rollback()

    statement, err := tx.Prepare(tableSQL)
    if err != nil {
           return err
    }
    defer statement.Close()

    _, err = statement.Exec()
    if err != nil {
        return err
    }
    return tx.Commit()
}

func _createSessionTable(ctx context.Context) error {
	tableSQL := `CREATE TABLE IF NOT EXISTS sessions (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        user_id TEXT NOT NULL,
        session_id TEXT NOT NULL,
        refresh_token TEXT NOT NULL,
        login_at TEXT NOT NULL,
        FOREIGN KEY (user_id) REFERENCES users (user_id)
	);`

	tx, err := DB.BeginTx(ctx, nil)
	if err != nil {
        return err
    }
    defer tx.Rollback()

    statement, err := tx.Prepare(tableSQL)
    if err != nil {
           return err
    }
    defer statement.Close()

    _, err = statement.Exec()
    if err != nil {
        return err
    }
    return tx.Commit()
}

func InsertUser(ctx context.Context, email string, username string, password string) error {

	query := `INSERT INTO users
	(user_id, email, username, created_at, password, sub_tier) VALUES
	(?, ?, ?, ?, ?, ?)`

	createdAt := time.Now().Unix()
    user_id, err := auth.GenUUID() 
    sub_tier := 0

    tx, err := DB.BeginTx(ctx, nil)
    if err != nil {
           return err
    }

    _, err = tx.ExecContext(ctx, query, user_id, email, username, createdAt, password, sub_tier)
    if err != nil {
        tx.Rollback()
        return err
    }
    return tx.Commit()
}

func GetUserByEmail(email string) (*User, error) {

	stmt, err := DB.Prepare("SELECT * FROM users WHERE email = ?")
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    var user User

    err = stmt.QueryRow(email).Scan(&user.Email, &user.UserName, &user.CreatedAt, &user.Password)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }

    return &user, nil
}
