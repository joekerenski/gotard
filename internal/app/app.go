package app

import (
    "gotard/internal/router"
    "gotard/internal/config"
    "gotard/internal/db"
    "database/sql"
    "path/filepath"
    "log"
    "os"
    "os/signal"
    "strings"
    "strconv"
    "syscall"
    "context"
    "net/http"
)

type App struct {
    Name    string
    Port    string
    Debug   bool 
    LoadConfig  bool    // change to path, ignore if empty
    DB      *sql.DB
    MainMux *router.Router
    Middlewares []func(http.Handler) http.Handler
}

func NewApp(name string, port string) *App {
    return &App{
        Name:    name,
        Port:    port,
        Debug:   false,
        LoadConfig:  false,
        DB:      nil,
        MainMux: router.NewRouter("MAIN"),
        Middlewares: []func(http.Handler) http.Handler{},
    }
}

func (app *App) Include(router *router.Router, prefix string) {
    app.MainMux.Mux.Handle(prefix+"/", http.StripPrefix(prefix, router))
}

func (app *App) AddMiddleware(middleware func(http.Handler) http.Handler) {
    app.Middlewares = append(app.Middlewares, middleware)
}

func (app *App) _applyMiddlewares(handler http.Handler) http.Handler {
    for i := len(app.Middlewares) -1; i >= 0; i-- {
        handler = app.Middlewares[i](handler)
    }
    return handler
}

func (app *App) ServeStaticFiles(htmlPath, assets string) {
    app.MainMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/" {
            http.ServeFile(w, r, filepath.Join(htmlPath, "index.html"))
            return
        }
        http.NotFound(w, r)
    })

    err := filepath.Walk(htmlPath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
            relPath, err := filepath.Rel(htmlPath, path)
            if err != nil {
                return err
            }
            urlPath := "/" + strings.TrimSuffix(relPath, ".html")

            app.MainMux.HandleFunc(urlPath, func(w http.ResponseWriter, r *http.Request) {
                app.MainMux.Logger.Printf("[%s] %s %s", app.MainMux.Tag, r.Method, r.URL.Path)
                http.ServeFile(w, r, path)
            })

            app.MainMux.Logger.Printf("Registered handler for %s -> %s", urlPath, path)
        } 
        return nil
    })

    if err != nil {
        log.Fatalf("Error walking the path %q: %v\n", htmlPath, err)
    }

    fs := http.FileServer(http.Dir(assets))
    app.MainMux.Handle("/assets/", http.StripPrefix("/assets/", fs))
    app.MainMux.Logger.Printf("Serving static assets from %s at /assets/", assets)
}

func (app *App) Run() {
    if app.LoadConfig {
        _ = config.LoadEnv("./local.env") 
        app.Name = config.GetEnv("APP_NAME", "Retardo")
        app.Port = config.GetEnv("APP_PORT", "6969")
        app.Debug, _ = strconv.ParseBool(config.GetEnv("DEBUG", "false"))
    }

    if app.Debug {
        config, _ := config.DumpConfigAsJSON("./local.env") 
        app.MainMux.Logger.Println("Loaded config.")
        app.MainMux.Logger.Printf("Current configuration:\n%s", config)
    }

    server := &http.Server{
        Addr:       ":" + app.Port,
        Handler:    app._applyMiddlewares(app.MainMux),
        ErrorLog:     app.MainMux.Logger,
    }
    
    var err error
    app.DB, err = db.InitDB(config.GetEnv("DB_NAME", "users.db"))
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
        return
    }
    defer app.DB.Close()

    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        sig := <- signalChan
        app.MainMux.Logger.Printf("Received signal: %s. Shutting down. Rip '%s' ... \n", sig, app.Name)
        server.Shutdown(context.Background())
    }()

    app.MainMux.Logger.Printf("%s is running on port %s\n", app.Name, app.Port)
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Error starting server: %v\n", err)
    }

    log.Println("Server stopped.")
}

