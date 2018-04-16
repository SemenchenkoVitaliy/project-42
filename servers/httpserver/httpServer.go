package httpserver

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/tcp"

	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
)

func tcpHandler(server tcp.Server) {
	err := server.Auth(tcp.AuthData{
		IP:   common.Config.HostIP,
		Port: common.Config.HostPort,
		Type: "http",
	})
	if err != nil {
		common.LogCritical(err, "unable to authentifacate")
		return
	}
	for {
		d, dt, e := server.Recieve()
		if e != nil {
			common.LogCritical(err, "unable to recieve a message from server")
			return
		}
		switch dt {
		case 0:
			continue
		case 4:
			var updCacheData tcp.UpdateCache
			err = json.Unmarshal(d, &updCacheData)
			if err != nil {
				common.Log(err, "encode data to update cache")
				continue
			}
			switch updCacheData.Product {
			case "manga":
				dbDriver.MangaCache.Remove(updCacheData.Name)
				if updCacheData.Pages {
					if updCacheData.PagesAll {
						dbDriver.MangaPagesCache.Remove(updCacheData.Name)
					} else {
						dbDriver.MangaPagesCache.RemoveChapter(updCacheData.Name, updCacheData.Chapter)
					}
				}
			case "ranobe":
			default:
			}
		default:
		}
	}
}

func openHttpServer() {
	r := mux.NewRouter()

	r.HandleFunc("/", httpMain).Methods("GET")
	r.HandleFunc("/manga", httpMangaMain).Methods("GET")
	r.HandleFunc("/manga/{name}", httpMangaInfo).Methods("GET")
	r.HandleFunc("/manga/{name}/{chapter}", httpMangaRead).Methods("GET")

	r.HandleFunc("/admin", httpAdmin).Methods("GET")
	r.HandleFunc("/admin/manga/{name}", httpAdminManga).Methods("GET")

	http.Handle("/", r)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	fmt.Printf("http server is opened on %v:%v\n", common.Config.HostIP, common.Config.HostPort)
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
