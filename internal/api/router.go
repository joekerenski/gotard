package api

import (
	"net/http"
    "os"
	"log"
)

type Router struct {
    mux *http.ServeMux
    mwAuth func(http.HandlerFunc) http.HandlerFunc
    mwExtra []func(http.HandlerFunc) http.HandlerFunc
    tag string
    protected bool
    logger *log.Logger
}

func NewRouter(tag string) *Router {
    return &Router{
        mux:        http.NewServeMux(),
        mwAuth:     nil,
        mwExtra:    []func(http.HandlerFunc) http.HandlerFunc{},
        tag:        tag,
        protected:  true,
        logger:     log.New(os.Stdout, "["+tag+"] ", log.LstdFlags),
    }
}

func(r *Router) Use(middleware func(http.HandlerFunc) http.HandlerFunc) {
    r.mwExtra = append(r.mwExtra, middleware)
}

func(r *Router) UseAuth(mwAuth func(http.HandlerFunc) http.HandlerFunc) {
    r.mwAuth = mwAuth 
}

func (r *Router) SetProtected(protected bool) {
    r.protected = protected
}

func (r *Router) Handle(method string, pattern string, handler http.HandlerFunc) {
    wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        if req.Method != method {
            http.Error(w, "Method not allowed you bish", http.StatusMethodNotAllowed)
            return
        }
        handler(w, req)
    })
    
    if r.mwAuth != nil && r.protected {
        wrappedHandler = r.mwAuth(wrappedHandler)
    }

    for i := len(r.mwExtra) -1; i >= 0; i-- {
        wrappedHandler = r.mwExtra[i](wrappedHandler)
    }

    r.mux.Handle(pattern, wrappedHandler)
}

func (r *Router)  ServeHTTP(w http.ResponseWriter, req *http.Request) {
    r.logger.Printf("%s %s", req.Method, req.URL.Path)
    r.mux.ServeHTTP(w, req)
}
