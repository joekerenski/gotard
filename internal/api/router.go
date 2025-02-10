package api

import (
	"net/http"
    "os"
	"log"
)

// TODO: switch to logging middleware so that files and api routes get logged
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
        Public:     []string{},
        Logger:     log.New(os.Stdout, "["+tag+"] ", log.LstdFlags),
    }
}

// TODO: MakePublic -> add path or pattern to the Router to disable mwAuth
// auth middleware looks for the public list
func (r *Router) MakePublic(pattern string) {
    r.Public = append(r.Public, pattern) 
}

func(r *Router) Use(middleware func(http.HandlerFunc) http.HandlerFunc) {
    r.MwExtra = append(r.MwExtra, middleware)
}

func(r *Router) UseAuth(mwAuth func(http.HandlerFunc) http.HandlerFunc) {
    r.MwAuth = mwAuth 
}

func (r *Router) Handle(pattern string, handler http.HandlerFunc) {
    
    if r.MwAuth != nil {
        handler = r.MwAuth(handler)
    }

    for i := len(r.MwExtra) -1; i >= 0; i-- {
        handler = r.MwExtra[i](handler)
    }

    r.Mux.Handle(pattern, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.Logger.Printf("%s %s", req.Method, req.URL.Path)
    r.Mux.ServeHTTP(w, req)
}


