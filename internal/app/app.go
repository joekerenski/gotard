package app

import (
    "gotard/internal/api"
    "gotard/internal/config"
    "gotard/internal/db"
    "database/sql"
    "log"
    "os"
    "os/signal"
    "strconv"
    "syscall"
    "context"
    "net/http"
)

type App struct {
    Name    string
    Port    string
    Debug   bool 
    Config  bool
    Context context.Context 
    Cancel  context.CancelFunc 
    DB      *sql.DB
    MainMux *http.ServeMux  // maybe make this a proper file server?
    Routers []*api.Router
    Middlewares []func(http.Handler) http.Handler
}

func NewApp(name string, port string) *App {
    ctx, cancel := context.WithCancel(context.Background()) // research better what this means
    return &App{
        Name:    name,
        Port:    port,
        Debug:   false,
        Config:  false,
        Context: ctx,
        Cancel:  cancel,
        DB:      nil,
        MainMux: http.NewServeMux(),
        Routers: []*api.Router{},
        Middlewares: []func(http.Handler) http.Handler{},
    }
}

func (app *App) Include(router *api.Router, prefix string) {
    app.Routers = append(app.Routers, router)
    app.MainMux.Handle(prefix+"/", http.StripPrefix(prefix, router))
}

func (app *App) AddMiddleware(middleware func(http.Handler) http.Handler) {
    app.Middlewares = append(app.Middlewares, middleware)
}

func (app *App) _applyMiddlewares(handler http.Handler) http.Handler {
    for _, middleware := range app.Middlewares {
        handler = middleware(handler)
    }
    return handler
}

func (app *App) Run() {
    server := &http.Server{
        Addr:       ":" + app.Port,
        Handler:    app._applyMiddlewares(app.MainMux),
    }
    
    if app.Config {
        _ = config.LoadEnv("./local.env") 
        app.Name = config.GetEnv("APP_NAME", "Retardo")
        app.Port = config.GetEnv("APP_PORT", "6969")
        app.Debug, _ = strconv.ParseBool(config.GetEnv("DEBUG", "false"))
    }

    if app.Debug {
        config, _ := config.DumpConfigAsJSON("./local.env") 
        log.Println("[DEBUG]: Loaded config.")
        log.Printf("[DEBUG]: Current configuration:\n%s", config)
    }

    db, err := db.InitDB(config.GetEnv("DB_NAME", "users.db"))
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer db.Close()

    app.Context = context.WithValue(app.Context, "DB", db)

    signalChan := make(chan os.Signal, 1)
    signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

    go func() {
        sig := <- signalChan
        log.Printf("Received signal: %s. Shutting down. Rip '%s' ... \n", sig, app.Name)
        app.Cancel()
        server.Shutdown(context.Background())
    }()

    log.Printf("%s is running on port %s\n", app.Name, app.Port)
    if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Error starting server: %v\n", err)
    }

    log.Println("Server stopped.")
}
