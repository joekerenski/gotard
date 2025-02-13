package main

import (
	"gotard/internal/app"
	"gotard/internal/middleware"
    "gotard/internal/router"
    "gotard/internal/db"
	"net/http"
    "encoding/json"
	// "fmt"
)

func main() {
    app := app.NewApp("Retardo", "8000")
    app.LoadConfig = true
    app.Debug = true
    app.Init()

    app.ServeStaticFiles("./static", "./static/assets")
    app.AddMiddleware(middleware.LoggingMiddleware)

    apiRouter := router.NewRouter("API")

    apiRouter.Use(middleware.PublicPathMiddleware(apiRouter.PublicPaths))
    apiRouter.MakePublic("/lolz")
    apiRouter.MakePublic("/login")
    apiRouter.Use(middleware.JWTMiddleware)
    apiRouter.Use(middleware.SessionMiddleware(app.DB))
    
    apiRouter.Handle("POST /login", router.LoginHandler(app.DB))

    apiRouter.HandleFunc("GET /lolz", func(w http.ResponseWriter, r *http.Request) {
        response := "Get them lolz fucking loser."
        w.Write([]byte(response))
    })

    apiRouter.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
        userInfo, ok := r.Context().Value("userInfo").(*db.User)
        if !ok {
            http.Error(w, "User info not found in context", http.StatusInternalServerError)
            return
        }

        err := json.NewEncoder(w).Encode(userInfo)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
    })

    app.Include(apiRouter, "/api")
    app.Run()
}

