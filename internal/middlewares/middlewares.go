package middlewares 

import (
    "gotard/internal/db"
    "gotard/internal/auth"
    "database/sql"
	"net/http"
	"context"
    "log"
)

type responseWriter struct {
    http.ResponseWriter
    statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        wrappedWriter := &responseWriter{w, http.StatusOK}

        next.ServeHTTP(wrappedWriter, r)

        log.Printf("[REQUEST] %s %s %d %s",
            r.Method,
            r.URL.Path,
            wrappedWriter.statusCode,
            http.StatusText(wrappedWriter.statusCode),
        )
    }
}

func JWTMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        cookie, err := r.Cookie("AuthToken")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        payload, err := auth.VerifyPayload(Secret, cookie.Value)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), "userId", payload.Sub)
        r = r.WithContext(ctx)

        next.ServeHTTP(w, r)
    }
}

func SessionMiddleware(dbConn *sql.DB, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userId, ok := r.Context().Value("userId").(string)
        if !ok {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        userInfo, err := db.GetUserById(userId, dbConn)
        if err != nil {
            http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            return
        }

        ctx := context.WithValue(r.Context(), "userInfo", userInfo)
        r = r.WithContext(ctx)

        next.ServeHTTP(w, r)
    }
}


