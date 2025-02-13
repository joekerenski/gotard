package router

import (
	"net/http"
    "os"
    "strings"
	"log"
)

type Router struct {
    Mux         *http.ServeMux
    Middlewares []func(http.Handler) http.Handler
    Tag         string
    PublicPaths map[string]bool
    Logger      *log.Logger
}

func NewRouter(tag string) *Router {
    return &Router{
        Mux:         http.NewServeMux(),
        Middlewares: []func(http.Handler) http.Handler{},
        Tag:         tag,
        PublicPaths:      make(map[string]bool),
        Logger:      log.New(os.Stdout, "["+tag+"] ", log.LstdFlags),
    }
}

func (r *Router) MakePublic(pattern string) {
    r.PublicPaths[pattern] = true
    r.Logger.Printf("Route %s marked as public", pattern)
}

func (r *Router) IsPublicPath(path string) bool {
    if r.PublicPaths[path] {
        return true
    }
    for pattern := range r.PublicPaths {
        if strings.HasSuffix(pattern, "/*") {
            prefix := strings.TrimSuffix(pattern, "/*")
            if strings.HasPrefix(path, prefix) {
                return true
            }
        }
    }
    return false
}


func (r *Router) Use(middleware func(http.Handler) http.Handler) {
    r.Middlewares = append(r.Middlewares, middleware)
}

func (r *Router) Handle(pattern string, handler http.Handler) {
    for i := len(r.Middlewares) - 1; i >= 0; i-- {
        handler = r.Middlewares[i](handler)
    }
    r.Mux.Handle(pattern, handler)
}

func (r *Router) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request)) {
    r.Handle(pattern, http.HandlerFunc(handler))
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.Logger.Printf("%s %s", req.Method, req.URL.Path)
    r.Mux.ServeHTTP(w, req)
}


