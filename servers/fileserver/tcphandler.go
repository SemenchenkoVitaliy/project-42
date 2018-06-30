package fileserver

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/netutils"
	"github.com/SemenchenkoVitaliy/project-42/utils"
)

func tcpHandler(server netutils.Server) {
	err := server.Auth(netutils.AuthData{
		IP:   utils.Config.IP,
		Port: utils.Config.Port,
		Type: "file",
		Id:   utils.Config.ServerId,
	})
	if err != nil {
		utils.LogCritical(err, "unable to authentifacate")
		return
	}
	for {
		data, dataType, err := server.Recieve()
		if err != nil {
			utils.LogCritical(err, "unable to recieve a message from server")
			return
		}
		switch dataType {
		case 0:
			continue
		case 2:
			var fileData netutils.WriteFileData
			err = json.Unmarshal(data, &fileData)
			if err != nil {
				utils.Log(err, "encode file to write")
				continue
			}
			if err = ioutil.WriteFile(utils.Config.SrcDir+"/"+fileData.Path, fileData.Data, 0777); err != nil {
				dir := utils.Config.SrcDir
				arr := strings.Split(fileData.Path, "/")
				for i := 0; i < len(arr)-1; i++ {
					dir += "/" + arr[i]
					os.Mkdir(dir, 0777)
				}
				if err = ioutil.WriteFile(utils.Config.SrcDir+"/"+fileData.Path, fileData.Data, 0777); err != nil {
					utils.Log(err, "write file: "+utils.Config.SrcDir+"/"+fileData.Path)
					continue
				}
			}
			ioutil.WriteFile(utils.Config.SrcDir+"/"+fileData.Path, fileData.Data, 0777)
		case 3:
			dir := string(data)
			err = os.Mkdir(utils.Config.SrcDir+dir, 0777)
			if err != nil {
				utils.Log(err, "create directory: "+dir)
				continue
			}
		default:
		}
	}

}

func tcpHandlerHashed(server netutils.Server) {
	err := server.Auth(netutils.AuthData{
		IP:   utils.Config.IP,
		Port: utils.Config.Port,
		Type: "file",
		Id:   utils.Config.ServerId,
	})
	if err != nil {
		utils.LogCritical(err, "unable to authentifacate")
		return
	}
	for {
		data, dataType, err := server.Recieve()
		if err != nil {
			utils.LogCritical(err, "unable to recieve a message from server")
			return
		}
		switch dataType {
		case 0:
			continue
		case 2:
			var fileData netutils.WriteFileData
			err = json.Unmarshal(data, &fileData)
			if err != nil {
				utils.Log(err, "encode file to write")
				continue
			}
			h := sha256.New()
			h.Write([]byte(fileData.Path))
			ioutil.WriteFile(utils.Config.SrcDir+"/"+base64.URLEncoding.EncodeToString(h.Sum(nil)), fileData.Data, 0777)
		case 3:
			dir := string(data)
			err = os.Mkdir(utils.Config.SrcDir+dir, 0777)
			if err != nil {
				utils.Log(err, "create directory: "+dir)
				continue
			}
		default:
		}
	}

}
