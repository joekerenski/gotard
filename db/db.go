package db

import (
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

    return nil
}

func _createUserTable(ctx context.Context) error {
	tableSQL := `CREATE TABLE IF NOT EXISTS users (
		email TEXT NOT NULL PRIMARY KEY,
	    username TEXT NOT NULL,
	    created_at INTEGER NOT NULL,
	    password TEXT NOT NULL
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
	(email, username, created_at, password) VALUES
	(?, ?, ?, ?)`

	createdAt := time.Now().Unix()

	tx, err := DB.BeginTx(ctx, nil)
    if err != nil {
           return err
    }

    _, err = tx.ExecContext(ctx, query, email, username, createdAt, password)
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
