package auth

import (
	"net/http"
	"context"
)

// TODO: also check refresh token if expired, and issue new jwt
func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("AuthToken")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        payload, err := VerifyPayload(Secret, cookie.Value)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), "userId", payload.Sub)
        r = r.WithContext(ctx)

        next.ServeHTTP(w, r)
    }
}

// func CreateUserSession(user_id string) (*db.Session, error) {
//
//     payload :=  NewPayload(user_id)
//     token, err :=  SignPayload(Secret, payload)
//     if err != nil {
//         log.Printf("ERROR: %s", err)
//         return nil, err
//         http.Error(w, "Error signing token", http.StatusInternalServerError)
//         return
//     }
//
//     ctx := context.Background()
//     refresh, err := db.InsertNewSession(ctx, user_id)
//     if err != nil {
//         log.Printf("ERROR: %s", err)
//         return nil, err
//         http.Error(w, "Error creating a session!", http.StatusInternalServerError)
//         return
//     }
//
//     log.Printf("The refresh token has been returned: %s", refresh)
//     return nil, err
//
//     http.SetCookie(w, &http.Cookie{
//         Name:     "AuthToken",
//         Value:    token,
//         Path:     "/",
//         HttpOnly: true,
//         Secure:   true,
//         SameSite: http.SameSiteStrictMode,
//         MaxAge:   3600,
//         Expires:  time.Now().Add(15 * time.Minute),
//     })
//     //
//     // http.SetCookie(w, &http.Cookie{
//     //     Name:     "RefreshToken",
//     //     Value:    token,
//     //     Path:     "/",
//     //     HttpOnly: true,
//     //     Secure:   true,
//     //     SameSite: http.SameSiteStrictMode,
//     //     MaxAge:   3600,
//     //     Expires:  time.Now().Add(24 * time.Hour),
//     // })
//
//     session = &Session{}
//     return session, err 
// }
