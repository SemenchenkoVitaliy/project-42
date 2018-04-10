package apiserver

import (
	"encoding/json"

	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/mangaLoader"
	"github.com/SemenchenkoVitaliy/project-42/tcp"
)

var mainServer tcp.Server

func tcpHandler(server tcp.Server) {
	mainServer = server
	mangaLoader.Init(server)

	err := server.Auth(tcp.AuthData{
		IP:   common.Config.HostIP,
		Port: common.Config.HostPort,
		Type: "api",
	})
	if err != nil {
		common.CreateLogCritical(err, "unable to authentifacate")
		return
	}
	for {
		_, dt, e := server.Recieve()
		if e != nil {
			common.CreateLogCritical(err, "unable to recieve a message from server")
			return
		}
		switch dt {
		case 0:
			continue
		default:
		}
	}
}

func WriteFile(path string, fileData []byte) {
	data := tcp.WriteFileData{
		Path: path,
		Data: fileData,
	}
	b, _ := json.Marshal(data)
	mainServer.Send(b, 2)
}

func MkDir(path string) {
	mainServer.Send([]byte(path), 3)
}
