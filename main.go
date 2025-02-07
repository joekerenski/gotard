package main

import (
	"gotard/internal/app"
	"gotard/internal/auth"
    "gotard/internal/api"
	"gotard/internal/api/v1"
	"net/http"
	"fmt"
)

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Printf("Request: %s %s\n", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

func main() {

    app := app.NewApp("Retardo", "8000")

    fs := http.FileServer(http.Dir("./static"))
    app.MainMux.HandleFunc("/", v1.IndexHandler)
    app.MainMux.HandleFunc("/home", auth.JWTMiddleware(v1.HomeHandler))

    apiRouter := api.NewRouter("API")

    apiRouter.SetProtected(false)
    apiRouter.UseAuth(auth.JWTMiddleware)
    apiRouter.Handle("POST", "/signup",  v1.SignupHandler)
    apiRouter.Handle("POST", "/login",   v1.LoginHandler)

    apiRouter.SetProtected(true)
    apiRouter.Handle("GET",  "/account", v1.AccountHandler)

    // init the database, too

    app.Include(apiRouter, "/api")
    app.AddMiddleware(loggingMiddleware)
    app.MainMux.Handle("/static/", http.StripPrefix("/static/", fs))
    app.Config = true
    app.Run()
}

