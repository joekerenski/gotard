package server

import (
	"gotard/auth"
	"gotard/db"
	"context"
	"net/http"
	"encoding/json"
	"log"
	"time"
	"fmt"
)

type Router struct {
    mux *http.ServeMux
    mwAuth func(http.HandlerFunc) http.HandlerFunc
    protected bool
    tag string
}

func NewRouter(middleware func(http.HandlerFunc) http.HandlerFunc, tag string) *Router {
    return &Router{
        mux:        http.NewServeMux(),
        mwAuth:     middleware,
        protected:  true,
        tag:        tag,
    }
}

func (r *Router) Handle(method string, pattern string, handler http.HandlerFunc) {
    wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        if req.Method != method {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }
        handler(w, req)
    })

    if r.protected {
        wrappedHandler = r.mwAuth(wrappedHandler)
    }

    r.mux.Handle(pattern, wrappedHandler)
}

func (r *Router) SetProtected(protected bool) {
    r.protected = protected
}

func (r *Router)  ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.mux.ServeHTTP(w, req)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./server/static/index.html")
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "./server/static/home.html")
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
	  	if err := r.ParseMultipartForm(1 << 20); err != nil {
            http.Error(w, "Unable to parse form", http.StatusBadRequest)
            return
	    }

		email := r.FormValue("email")
        username := r.FormValue("username")
        password := r.FormValue("password")

        hashedPassword, err := auth.HashPassword(password)
        if err != nil {
            http.Error(w, "Error hashing password", http.StatusInternalServerError)
            return
        }

        ctx := context.Background()
        if err := db.InsertUser(ctx, email, username, hashedPassword); err != nil {
            http.Error(w, "Error inserting user into database", http.StatusInternalServerError)
            return
        }

        response := "User registered successfully!"
        w.Write([]byte(response))
    } else {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func loginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodPost {
	  	if err := r.ParseMultipartForm(1 << 20); err != nil {
            http.Error(w, "Unable to parse form", http.StatusBadRequest)
            return
	    }

		email := r.FormValue("email")
        //username := r.FormValue("username")
        password := r.FormValue("password")

        user, err := db.GetUserByEmail(email)
        if err != nil {
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

        payload := auth.NewPayload(email)
        token, err := auth.SignPayload(auth.Secret, payload)
        if err != nil {
            http.Error(w, "Error signing token", http.StatusInternalServerError)
            return
        }
    
        // TODO: how do we define a refresh token? and then define a session struct?
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

        // response := "Login successful!"
        // w.Write([]byte(response))
        http.Redirect(w, r, "/home", http.StatusSeeOther)
        return

    } else {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
    }
}

func accountHandler(w http.ResponseWriter, r *http.Request) {
    userEmail := r.Context().Value("userEmail").(string)

    userData, err := db.GetUserByEmail(userEmail)
    if err != nil {
	    http.Error(w, "Error retrieving user data!", http.StatusInternalServerError)
		log.Printf("ERROR occurred: %s", err)
	    return
    }

    userJSON, err := json.Marshal(userData)
    if err != nil {
	    http.Error(w, "Could not serialize user data!", http.StatusInternalServerError)
		log.Printf("ERROR occurred: %s", err)
	    return
    }

    w.Write([]byte(userJSON))
}

func StartServer() {
    
    mainMux := http.NewServeMux()
    
    fs := http.FileServer(http.Dir("./server/static"))
    mainMux.Handle("/static/", http.StripPrefix("/static/", fs))
    mainMux.HandleFunc("/", indexHandler)
    mainMux.HandleFunc("/home", auth.JWTMiddleware(homeHandler))

    apiRouter := NewRouter(auth.JWTMiddleware, "api")

    apiRouter.SetProtected(false)
    apiRouter.Handle("POST", "/signup",  signupHandler)
    apiRouter.Handle("POST", "/login",   loginHandler)

    apiRouter.SetProtected(true)
    apiRouter.Handle("GET",  "/account", accountHandler)
    
    mainMux.Handle("/api/", http.StripPrefix("/api", apiRouter))

	fmt.Println("Server is running on :8000")
	if err := http.ListenAndServe(":8000", mainMux); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

