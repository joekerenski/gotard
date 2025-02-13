package main

import (
    "net/http" 
    "log"
    "fmt"
    "os"
)

type Router struct {
    Mux *http.ServeMux
    MwAuth func(http.HandlerFunc) http.HandlerFunc
    MwExtra []func(http.HandlerFunc) http.HandlerFunc
    Tag string
    Public []string
    Logger *log.Logger
}

func NewRouter(tag string) *Router {
    return &Router{
        Mux:        http.NewServeMux(),
        MwAuth:     nil,
        MwExtra:    []func(http.HandlerFunc) http.HandlerFunc{},
        Tag:        tag,
        Public:  []string{},
        Logger:     log.New(os.Stdout, "["+tag+"] ", log.LstdFlags),
    }
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.Logger.Printf("%s %s", req.Method, req.URL.Path)
    r.Mux.ServeHTTP(w, req)
}

func main() {
    
    fileMux := http.NewServeMux()
    apiMux := http.NewServeMux()

    // fileMux
    fs := http.FileServer(http.Dir("./static"))
    
    // fileRouter.Mux.Handle("GET /", fs)
    fileMux.Handle("GET /", fs)

    fileMux.HandleFunc("GET /home", func(w http.ResponseWriter, r *http.Request) {
        http.ServeFile(w, r, "./static/home.html")
    })
    
    // apiMux
    apiMux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
        response := "<h1>We suck at software engineering!<h1>"
        w.Write([]byte(response))
    })
    
    apiMux.HandleFunc("GET /user/{id}", func(w http.ResponseWriter, r *http.Request) {
        id := r.PathValue("id")
        response := fmt.Sprintf("<h1>Any path will do, I don't mind: %s.<h1>", id)
        w.Write([]byte(response))
    })

    fileMux.Handle("GET /api/", http.StripPrefix("/api", apiMux))

    log.Printf("We is running on port 8000. We waz kangs.")
    if err := http.ListenAndServe(":8000", fileMux); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Error starting server: %v\n", err)
    }
}
