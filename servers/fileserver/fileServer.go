package fileserver

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/tcp"
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

func tcpHandler(server tcp.Server) {
	err := server.Auth(tcp.AuthData{
		IP:   common.Config.HostIP,
		Port: common.Config.HostPort,
		Type: "file",
	})
	if err != nil {
		common.CreateLogCritical(err, "unable to authentifacate")
		return
	}
	for {
		d, dt, e := server.Recieve()
		if e != nil {
			common.CreateLogCritical(err, "unable to recieve a message from server")
			return
		}
		switch dt {
		case 0:
			continue
		case 2:
			var fileData tcp.WriteFileData
			err = json.Unmarshal(d, &fileData)
			if err != nil {
				common.CreateLog(err, "encode file to write")
				continue
			}
			ioutil.WriteFile(common.Config.SrcDir+"/"+fileData.Path, fileData.Data, 0777)
		case 3:
			dir := string(d)
			err = os.Mkdir(dir, 0777)
			if err != nil {
				common.CreateLog(err, "create directory: "+dir)
				continue
			}
		default:
		}
	}
}

func openHttpServer() {
	fs := http.FileServer(http.Dir(common.Config.SrcDir))
	http.Handle("/", noDirListing(fs))

	fmt.Printf("file server is opened on %v:%v\n", common.Config.HostIP, common.Config.HostPort)
	err := http.ListenAndServe(fmt.Sprintf("%v:%v", common.Config.HostIP, common.Config.HostPort), nil)
	if err != nil {
		common.CreateLogCritical(err, fmt.Sprintf("open http server on %v:%v\n", common.Config.HostIP, common.Config.HostPort))
	}

}

func Start() {
	go openHttpServer()

	cert, err := tls.LoadX509KeyPair("certs/cert.pem", "certs/key.pem")
	if err != nil {
		common.CreateLogCritical(err, "load X509 key pair")
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%v:%v", common.Config.Tcp.HostIP, common.Config.Tcp.HostPort), &config)
	if err != nil {
		common.CreateLogCritical(err, "connect through tcp to main server")
	}

	server := tcp.Server{}
	server.Start(conn, tcpHandler)
}
