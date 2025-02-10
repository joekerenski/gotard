package middlewares 

import (
    "gotard/internal/auth"
	"net/http"
    "time"
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

        log.Printf("[%s] %s %s %d %s",
            time.Now().Format("2006-01-02 15:04:05"),
            r.Method,
            r.URL.Path,
            wrappedWriter.statusCode,
            http.StatusText(wrappedWriter.statusCode),
        )
    })
}

// // use slog here instead
// func LoggingMiddleware(next http.Handler) http.Handler {
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         fmt.Printf("Request: %s %s\n", r.Method, r.URL.Path)
//         next.ServeHTTP(w, r)
//     })
// }

// for api auth: take userID from previous middleware and extract userInfo + sessionInfo, + child Context
func SessionMiddleware(next http.HandlerFunc) http.HandlerFunc {
    // check context first: userID in there? 
    // then pull user & sessin info and fill context

    return func(w http.ResponseWriter, r *http.Request) {
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
        r = r.WithContext(ctx)

        next.ServeHTTP(w, r)
    }
}
