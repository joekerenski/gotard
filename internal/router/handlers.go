package router

import (
    "gotard/internal/auth"
    "gotard/internal/db"
    "database/sql"
    "time"
    "log"
    "net/http"
)
//
// func SignupHandler(w http.ResponseWriter, r *http.Request) {
// 		email := r.FormValue("email")
//         username := r.FormValue("username")
//         password := r.FormValue("password")
//
//         hashedPassword, err := auth.HashPassword(password)
//         if err != nil {
//             http.Error(w, "Error hashing password", http.StatusInternalServerError)
//             return
//         }
//
//         if err := db.InsertUser(email, username, hashedPassword); err != nil {
//             http.Error(w, "Error inserting user into database", http.StatusInternalServerError)
//             return
//         }
//
//         response := "User registered successfully!"
//         w.Write([]byte(response))
// }
// login info ->  handler gets a conn -> make db query -> var

func LoginHandler(conn *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
		email := r.FormValue("email")
        password := r.FormValue("password")

        user, err := db.GetUserByField("email", email, conn)
        if err != nil {
            log.Printf("Error executing transaction: %v", err)
            http.Error(w, "Error retrieving user", http.StatusInternalServerError)
            return
        }

        if user == nil {
            http.Error(w, "User not found", http.StatusNotFound)
            return
        }

        if !auth.VerifyPassword(user.Password, password) {
	        http.Error(w, "Invalid password", http.StatusUnauthorized)
	        return
        }

        payload := auth.NewPayload(user.Id)
        token, err := auth.SignPayload(auth.Secret, payload)
        if err != nil {
            http.Error(w, "Error signing token", http.StatusInternalServerError)
            return
        }

        http.SetCookie(w, &http.Cookie{
            Name:     "AuthToken",
            Value:    token,
            Path:     "/",
            HttpOnly: true,
            Secure:   true,
            SameSite: http.SameSiteStrictMode,
            MaxAge:   3600,
            Expires:  time.Now().Add(1 * time.Minute),
        })

        sessionStr, err := db.CreateSessionID()
        refresh, err := auth.CreateRefreshToken(sessionStr, auth.RefreshSecret)

        _,  err = db.InsertNewSession(user.Id, refresh, sessionStr, conn)
        if err != nil {
            http.Error(w, "Error creating a session!", http.StatusInternalServerError)
            return
        }

        http.SetCookie(w, &http.Cookie{
            Name:     "RefreshToken",
            Value:    refresh,
            Path:     "/",
            HttpOnly: true,
            Secure:   true,
            SameSite: http.SameSiteStrictMode,
            MaxAge:   3600,
            Expires:  time.Now().Add(24 * time.Hour),
        })

        response := "Login successful!"
        w.Write([]byte(response))
    }
}
    //
// func AccountHandler(w http.ResponseWriter, r *http.Request) {
//     user_id := r.Context().Value("userId").(string)
//
//     userData, err := db.GetUserById(user_id)
//     if err != nil {
// 	    http.Error(w, "Error retrieving user data!", http.StatusInternalServerError)
// 		log.Printf("ERROR occurred: %s", err)
// 	    return
//     }
//
//     userJSON, err := json.Marshal(userData)
//     if err != nil {
// 	    http.Error(w, "Could not serialize user data!", http.StatusInternalServerError)
// 		log.Printf("ERROR occurred: %s", err)
// 	    return
//     }
//
//     w.Write([]byte(userJSON))
// }
