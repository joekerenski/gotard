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
    ID            string    `json:"id"`
    Email         string    `json:"email"`
    UserName      string    `json:"username"`
    CreatedAt     time.Time `json:"created_at"`
    Password      string    `json:"-"`
    SubTier       int       `json:"sub_tier"`
    LastLoginAt   time.Time `json:"last_login_at"`
    IsActive      bool      `json:"is_active"`
    Role          string    `json:"role"`  // admin, authenticated, unchecked
    EmailVerified bool      `json:"email_verified"`
}

type Session struct {
    ID           string    `json:"id"`
    UserID       string    `json:"user_id"`
    SessionID    string    `json:"session_id"`
    RefreshToken string    `json:"-"`
    LoginAt      time.Time `json:"login_at"`
    ExpiresAt    time.Time `json:"expires_at"`
    LastUsedAt   time.Time `json:"last_used_at"`
    IsActive     bool      `json:"is_active"`
    UserAgent    string    `json:"user_agent"`
    IPAddress    string    `json:"ip_address"`
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
	    created_at TIMESTAMP NOT NULL,
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
        login_at TIMESTAMP NOT NULL,
        expires_at TIMESTAMP NOT NULL,
        last_used_at TIMESTAMP NOT NULL,
        is_active BOOLEAN DEFAULT TRUE,
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
        log.Printf("Error executing transaction: %v", err)
        return err
    }
    return tx.Commit()
}

func GetUserByEmail(email string) (*User, error) {

	stmt, err := DB.Prepare("SELECT user_id, email, username, created_at, password FROM users WHERE email = ?")
    if err != nil {
        return nil, err
    }
    defer stmt.Close()

    var user User

    err = stmt.QueryRow(email).Scan(&user.Id, &user.Email, &user.UserName, &user.CreatedAt, &user.Password)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &user, nil
}

func InsertNewSession(ctx context.Context, user_id string) error {

    sessionStr, err := auth.CreateSessionID()
    refresh, err := auth.CreateRefreshToken(sessionStr, auth.RefreshSecret)
 
    now := time.Now()
    expires_at := now.Add(24 * time.Hour)

	query := `INSERT INTO sessions
	(user_id, session_id, refresh_token, login_at, expires_at, last_used_at) VALUES
	(?, ?, ?, ?, ?, ?)`

    tx, err := DB.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    _, err = tx.ExecContext(ctx, query, user_id, sessionStr, refresh, now, expires_at, now)
    if err != nil {
        log.Printf("Error executing transaction: %v", err)
        return err
    }
    return tx.Commit()
}
