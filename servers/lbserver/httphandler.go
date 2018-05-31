package lbserver

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/SemenchenkoVitaliy/project-42/netutils"
)

func root(w http.ResponseWriter, r *http.Request) {
	worker, ok := httpServers.GetOne()
	if !ok {
		netutils.InternalError(w, nil, "All http servers are down")
		return
	}

	u, err := url.Parse(fmt.Sprintf("http://%v:%v", worker.Info.IP, worker.Info.Port))
	if err != nil {
		netutils.InternalError(w, err, "All http servers are down")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, r)

}

func api(w http.ResponseWriter, r *http.Request) {
	worker, ok := apiServers.GetOne()
	if !ok {
		netutils.InternalError(w, nil, "All api servers are down")
		return
	}

	u, err := url.Parse(fmt.Sprintf("http://%v:%v", worker.Info.IP, worker.Info.Port))
	if err != nil {
		netutils.InternalError(w, err, "All api servers are down")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, r)
}

func file(w http.ResponseWriter, r *http.Request) {
	ids, err := db.FindEntry(r.URL.Path)
	if err != nil {
		netutils.NotFoundError(w, err, "find fs entry: "+r.URL.Path)
		return
	}
	worker, ok := fileServers.GetOneFrom(ids)
	// worker, ok := fileServers.GetOne()
	if !ok {
		netutils.InternalError(w, nil, "All file servers are down")
		return
	}

	u, err := url.Parse(fmt.Sprintf("http://%v:%v", worker.Info.IP, worker.Info.Port))
	if err != nil {
		netutils.InternalError(w, err, "All file servers are down")
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ServeHTTP(w, r)
}
