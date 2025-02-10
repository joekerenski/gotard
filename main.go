package main

import (
	"gotard/internal/app"
	// "gotard/internal/middlewares"
    "gotard/internal/api"
	// "gotard/internal/api/v1"
	"net/http"
	// "fmt"
)

func main() {
    app := app.NewApp("Retardo", "8000")
    app.ServeStaticFiles("./static", "./static/assets")
    // app.AddMiddleware(middlewares.LoggingMiddleware)

    apiRouter := api.NewRouter("API")

    // apiRouter.SetProtected(false)
    // apiRouter.UseAuth(auth.JWTMiddleware)
    // apiRouter.Handle("POST /signup",  v1.SignupHandler)
    // apiRouter.Handle("POST /login",   v1.LoginHandler)

    // apiRouter.SetProtected(true)
    // apiRouter.Handle("GET",  "/account", v1.AccountHandler)
    apiRouter.Handle("GET /lolz", func(w http.ResponseWriter, r *http.Request) {
        response := "Get them lolz fucking loser."
        w.Write([]byte(response))
    })
    app.Include(apiRouter, "/api")
    // app.AddMiddleware(loggingMiddleware)
    app.LoadConfig = false
    app.Run()
}

