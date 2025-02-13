package router

// import (
//     "gotard/internal/auth"
//     "gotard/internal/db"
//     "time"
//     "encoding/json"
//     "log"
//     "context"
//     "net/http"
// )
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
//
// func LoginHandler(w http.ResponseWriter, r *http.Request) {
//         // use both mail and username to be unique
// 		email := r.FormValue("email")
//         // username := r.FormValue("username")
//         password := r.FormValue("password")
//
//         // fix this: get user via email, create jwt with user_id
//         user, err := db.GetUserById(email)
//         if err != nil {
//             log.Printf("Error executing transaction: %v", err)
//             http.Error(w, "Error retrieving user", http.StatusInternalServerError)
//             return
//         }
//
//         if user == nil {
//             http.Error(w, "User not found", http.StatusNotFound)
//             return
//         }
//
//         if !auth.VerifyPassword(user.Password, password) {
// 	        http.Error(w, "Invalid password", http.StatusUnauthorized)
// 	        return
//         }
//
//         // session := CreateNewSession(user.Id)
//
//         payload := auth.NewPayload(user.Id)
//         token, err := auth.SignPayload(auth.Secret, payload)
//         if err != nil {
//             http.Error(w, "Error signing token", http.StatusInternalServerError)
//             return
//         }
//
//         http.SetCookie(w, &http.Cookie{
//             Name:     "AuthToken",
//             Value:    token,
//             Path:     "/",
//             HttpOnly: true,
//             Secure:   true,
//             SameSite: http.SameSiteStrictMode,
//             MaxAge:   3600,
//             Expires:  time.Now().Add(1 * time.Minute),
//         })
//
//         sessionStr, err := db.CreateSessionID()
//         refresh, err := auth.CreateRefreshToken(sessionStr, auth.RefreshSecret)
//
//         ctx := context.Background()
//         _,  err = db.InsertNewSession(user.Id, refresh, sessionStr)
//         if err != nil {
//             http.Error(w, "Error creating a session!", http.StatusInternalServerError)
//             return
//         }
//
//         http.SetCookie(w, &http.Cookie{
//             Name:     "RefreshToken",
//             Value:    refresh,
//             Path:     "/",
//             HttpOnly: true,
//             Secure:   true,
//             SameSite: http.SameSiteStrictMode,
//             MaxAge:   3600,
//             Expires:  time.Now().Add(24 * time.Hour),
//         })
//
//         response := "Login successful!"
//         w.Write([]byte(response))
// }
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
