package apiserver

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/SemenchenkoVitaliy/project-42/common"
	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
	"github.com/SemenchenkoVitaliy/project-42/tcp"
)

func openHttpServer() {
	r := mux.NewRouter()

	r.HandleFunc("/", apiGetMain).Methods("GET")
	r.HandleFunc("/", apiCmd).Methods("POST")

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
		common.LogCritical(err, fmt.Sprintf("open http server on %v:%v\n", common.Config.HostIP, common.Config.HostPort))
	}

}

func Start() {
	dbDriver.Init()
	go openHttpServer()

	cert, err := tls.LoadX509KeyPair("certs/cert.pem", "certs/key.pem")
	if err != nil {
		common.LogCritical(err, "load X509 key pair")
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%v:%v", common.Config.Tcp.HostIP, common.Config.Tcp.HostPort), &config)
	if err != nil {
		common.LogCritical(err, "connect through tcp to main server")
	}

	server := tcp.Server{}
	server.Start(conn, tcpHandler)
}
