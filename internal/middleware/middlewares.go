package middleware

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

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        wrappedWriter := &responseWriter{w, http.StatusOK}
        next.ServeHTTP(wrappedWriter, r)
        log.Printf("[REQUEST] %s %s %d %s",
            r.Method,
            r.URL.Path,
            wrappedWriter.statusCode,
            http.StatusText(wrappedWriter.statusCode),
        )
    })
}

func JWTMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        isPublic, _ := r.Context().Value("isPublicRoute").(bool)
        log.Printf("Checking the public route: %s is %s", r.URL.Path, isPublic)
        if isPublic {
            next.ServeHTTP(w, r)
            return
        }

        cookie, err := r.Cookie("AuthToken")
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        payload, err := auth.VerifyPayload(auth.Secret, cookie.Value)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), "userId", payload.Sub)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func SessionMiddleware(dbConn *sql.DB) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userId, ok := r.Context().Value("userId").(string)
            if !ok {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            userInfo, err := db.GetUserByField("user_id", userId, dbConn)
            if err != nil {
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
                return
            }
            
            log.Printf("Extracted user info: %s", userInfo)
            ctx := context.WithValue(r.Context(), "userInfo", userInfo)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func PublicPathMiddleware(publicPaths map[string]bool) func(next http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        isPublic := publicPaths[r.URL.Path]
        ctx := context.WithValue(r.Context(), "isPublicRoute", isPublic)
        next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
