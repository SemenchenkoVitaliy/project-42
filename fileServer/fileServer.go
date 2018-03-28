package fileServer

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/common"
)

func noDirListing(h http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("403 - access forbidden"))
			return
		}
		h.ServeHTTP(w, r)
	})
}

func Start() {
	go openTCPServer()

	fs := http.FileServer(http.Dir(common.Config.SrcDir))
	http.Handle("/", noDirListing(fs))

	fmt.Printf("file server is opened on %v:%v\n", common.Config.Ftp.Host, common.Config.Ftp.Port)

	http.ListenAndServe(fmt.Sprintf("%v:%v", common.Config.Ftp.Host, common.Config.Ftp.Port), nil)
}
