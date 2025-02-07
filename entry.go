package main

import (
	"gotard/internal/app"
	"gotard/internal/auth"
	"gotard/internal/api"
	"gotard/internal/api/v1"
	"net/http"
	"encoding/json"
    "os"
	"log"
	"time"
	"fmt"
)

func main() {
    
    mainMux := http.NewServeMux()
    
    fs := http.FileServer(http.Dir("./server/static"))
    mainMux.Handle("/static/", http.StripPrefix("/static/", fs))
    mainMux.HandleFunc("/", indexHandler)
    mainMux.HandleFunc("/home", auth.JWTMiddleware(homeHandler))

    apiRouter := NewRouter("api")

    apiRouter.SetProtected(false)
    apiRouter.UseAuth(auth.JWTMiddleware)
    apiRouter.Handle("POST", "/signup",  signupHandler)
    apiRouter.Handle("POST", "/login",   loginHandler)

    apiRouter.SetProtected(true)
    apiRouter.Handle("GET",  "/account", accountHandler)
    
    mainMux.Handle("/api/", http.StripPrefix("/api", apiRouter))

    // TODO: set `protected` via the app, easier that way. let it handle the routers.
    app := app.NewApp("Retardo", "8000")

	fmt.Println("Server is running on :8000")
	if err := http.ListenAndServe(":8000", mainMux); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

