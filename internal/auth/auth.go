package auth

import (
	"net/http"
	"context"
)

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
