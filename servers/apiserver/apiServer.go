package apiserver

import (
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/tcp"
)

func openHttpServer() {
	r := mux.NewRouter()

	r.HandleFunc("/", apiGetMain).Methods("GET")

	r.HandleFunc("/manga", apiGetMangaMain).Methods("GET")
	r.HandleFunc("/manga", apiChangeMangaMain).Methods("POST")

	r.HandleFunc("/manga/{name}", apiGetMangaInfo).Methods("GET")
	r.HandleFunc("/manga/{name}", apiChangeMangaInfo).Methods("POST")

	r.HandleFunc("/manga/{name}/{chapter}", apiGetMangaRead).Methods("GET")
	r.HandleFunc("/manga/{name}/{chapter}", apiChangeMangaChapter).Methods("POST")

	http.Handle("/", r)

	fmt.Printf("api server is opened on %v:%v\n", common.Config.HostIP, common.Config.HostPort)
	err := http.ListenAndServe(fmt.Sprintf("%v:%v", common.Config.HostIP, common.Config.HostPort), nil)
	if err != nil {
		common.CreateLogCritical(err, fmt.Sprintf("open http server on %v:%v\n", common.Config.HostIP, common.Config.HostPort))
	}

}

func Start() {
	go openHttpServer()

	conn, err := net.Dial("tcp", fmt.Sprintf("%v:%v", common.Config.Tcp.HostIP, common.Config.Tcp.HostPort))
	if err != nil {
		common.CreateLogCritical(err, "connect through tcp to main server")
	}
	server := tcp.Server{}
	server.Start(conn, tcpHandler)
}
