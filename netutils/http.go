package netutils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/utils"
	"github.com/gorilla/mux"
)

/******************************************************************************
	Http router
******************************************************************************/

type route struct {
	Path    string
	Method  string
	Handler http.HandlerFunc
}

type HTTPServer struct {
	routes      []route
	dirs        map[string]string
	handleFuncs map[string]http.HandlerFunc
}

func NewHTTPServer() (s *HTTPServer) {
	return &HTTPServer{
		routes:      []route{},
		dirs:        make(map[string]string),
		handleFuncs: make(map[string]http.HandlerFunc),
	}
}

func (s *HTTPServer) AddRoute(path, method string, handler http.HandlerFunc) {
	s.routes = append(s.routes, route{
		Path:    path,
		Method:  method,
		Handler: handler,
	})
}

func (s *HTTPServer) Handle(path string, handler http.HandlerFunc) {
	s.handleFuncs[path] = handler
}

func (s *HTTPServer) AddDir(path, dir string) {
	s.dirs[path] = dir
}

func (s *HTTPServer) Listen(ip string, port int) {
	if len(s.handleFuncs) != 0 {

		for path, handler := range s.handleFuncs {
			http.Handle(path, handler)
		}
	} else {
		router := mux.NewRouter()
		for _, r := range s.routes {
			router.HandleFunc(r.Path, r.Handler).Methods(r.Method)
		}
		for path, dir := range s.dirs {
			http.Handle(path, http.StripPrefix(path, http.FileServer(http.Dir(dir))))
		}
		http.Handle("/", router)
	}
	hostname := fmt.Sprintf("%v:%v", ip, port)
	fmt.Println("http server is opened on " + hostname)
	if err := http.ListenAndServe(hostname, nil); err != nil {
		utils.LogCritical(err, "open http server on "+hostname)
	}
}

/******************************************************************************
	Domain router
******************************************************************************/

type DomainRouter struct {
	subdomains map[string]http.Handler
	main       http.Handler
}

func NewDomainRouter(handler http.HandlerFunc) (dr *DomainRouter) {
	return &DomainRouter{
		subdomains: make(map[string]http.Handler),
		main:       handler,
	}
}

func (dr *DomainRouter) AddSubdomain(subdomain string, handler http.HandlerFunc) {
	dr.subdomains[subdomain] = handler
}

func (dr DomainRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domainParts := strings.Split(r.Host, ".")

	if len(domainParts) == 1 {
		dr.main.ServeHTTP(w, r)
	} else if handler := dr.subdomains[domainParts[0]]; handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		NotFoundError(w, nil, "access subdomain "+domainParts[0])
	}
}

func (dr DomainRouter) Listen(ip string, port int) {
	hostname := fmt.Sprintf("%v:%v", ip, port)
	fmt.Println("Main server is being opened on " + hostname)
	if err := http.ListenAndServe(hostname, dr); err != nil {
		utils.LogCritical(err, "open http server on "+hostname)
	}
}

/******************************************************************************
	Error handlers
******************************************************************************/

func InternalError(w http.ResponseWriter, err error, text string) {
	if err != nil {
		utils.Log(err, text)
	}
	http.Error(w, "Error happend: "+text, http.StatusInternalServerError)
}

func NotFoundError(w http.ResponseWriter, err error, text string) {
	http.Error(w, "Error happend: "+text, http.StatusNotFound)
}

func ForbiddenError(w http.ResponseWriter, err error, text string) {
	http.Error(w, "Error happend: "+text, http.StatusForbidden)
}

func UnauthorizedError(w http.ResponseWriter, err error, text string) {
	http.Error(w, "Error happend: "+text, http.StatusUnauthorized)
}
