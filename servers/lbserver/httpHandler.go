package lbserver

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/common"
)

func openHttpServer() {
	fsMux := http.NewServeMux()
	apiMux := http.NewServeMux()
	httpMux := http.NewServeMux()

	fsMux.HandleFunc("/", fsHandler)
	apiMux.HandleFunc("/", apiHandler)
	httpMux.HandleFunc("/", httpHandler)

	var mux muxCustom
	mux.Main(httpMux)
	mux.Subdomains("img", fsMux)
	mux.Subdomains("api", apiMux)

	hostname := fmt.Sprintf("%v:%v", common.Config.HostIP, common.Config.HostPort)
	fmt.Printf("Main server is opened on %v\n", hostname)

	if err := http.ListenAndServe(hostname, mux); err != nil {
		common.LogCritical(err, "open main http server on "+hostname)
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	worker, err := httpServers.GetOne()
	if err != nil {
		http.Error(w, "All http servers are down", 500)
		return
	}

	u, err := url.Parse(fmt.Sprintf("http://%v:%v", worker.IP, worker.Port))
	if err != nil {
		http.Error(w, "All http servers are down", 500)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, r)
}
func apiHandler(w http.ResponseWriter, r *http.Request) {
	worker, err := apiServers.GetOne()
	if err != nil {
		http.Error(w, "All api servers are down", 500)
		return
	}

	u, err := url.Parse(fmt.Sprintf("http://%v:%v", worker.IP, worker.Port))
	if err != nil {
		http.Error(w, "All api servers are down", 500)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, r)
}
func fsHandler(w http.ResponseWriter, r *http.Request) {
	worker, err := fileServers.GetOne()
	if err != nil {
		http.Error(w, "All file servers are down", 500)
		return
	}

	u, err := url.Parse(fmt.Sprintf("http://%v:%v", worker.IP, worker.Port))
	if err != nil {
		http.Error(w, "All file servers are down", 500)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, r)
}

type muxCustom struct {
	subdomains map[string]http.Handler
	main       http.Handler
}

func (m *muxCustom) Main(handler http.Handler) {
	m.subdomains = make(map[string]http.Handler)
	m.main = handler
}

func (m *muxCustom) Subdomains(subdomain string, handler http.Handler) {
	m.subdomains[subdomain] = handler
}

func (m muxCustom) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domainParts := strings.Split(r.Host, ".")
	var handler http.Handler

	if len(domainParts) == 1 {
		m.main.ServeHTTP(w, r)
	} else if handler = m.subdomains[domainParts[0]]; handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		http.Error(w, "Not found", 404)
	}
}
