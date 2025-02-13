package main

import (
	"gotard/internal/app"
	"gotard/internal/middleware"
    "gotard/internal/router"
	"net/http"
	// "fmt"
)

// AddMiddleware + PublicPaths are handled by App globally, DONE
// somehow use the ServeMux matching to whitelist routes
// Routers dont need separate middleware, App keeps track and applies them all
// so then middleware can use appstate (incl. requesting a db conn.)

// app.AddMiddleware(blah) // for all sub routers too 
// app.MakePublic(pattern string) // pass in pre-stripped path
//
// apiRouter := router.NewRouter("Tag")
// ...
// app.Include(apiRouter, "/api") // what prefix to use /user -> /api/user

func main() {
    app := app.NewApp("Retardo", "8000")
    app.ServeStaticFiles("./static", "./static/assets")
    app.AddMiddleware(middleware.LoggingMiddleware)

    apiRouter := router.NewRouter("API")

    apiRouter.MakePublic("/lolz")

    // apiRouter.Use(middleware.PublicPathMiddleware(apiRouter.PublicPaths))
    // apiRouter.Use(middleware.JWTMiddleware)

    apiRouter.HandleFunc("GET /lolz", func(w http.ResponseWriter, r *http.Request) {
        response := "Get them lolz fucking loser."
        w.Write([]byte(response))
    })

    apiRouter.HandleFunc("GET /user", func(w http.ResponseWriter, r *http.Request) {
        userId, _ := r.Context().Value("userId").(string)
        w.Write([]byte(userId))
    })

    app.Include(apiRouter, "/api")
    app.LoadConfig = true
    app.Debug = true
    app.Run()
}

