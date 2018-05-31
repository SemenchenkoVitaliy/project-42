package httpserver

import (
	"encoding/json"

	"github.com/SemenchenkoVitaliy/project-42/netutils"
	"github.com/SemenchenkoVitaliy/project-42/utils"
)

func tcpHandler(server netutils.Server) {
	err := server.Auth(netutils.AuthData{
		IP:   utils.Config.IP,
		Port: utils.Config.Port,
		Type: "http",
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
		case 4:
			var updCacheData netutils.UpdateCache
			err = json.Unmarshal(data, &updCacheData)
			if err != nil {
				utils.Log(err, "encode data to update cache")
				continue
			}
			switch updCacheData.Product {
			case "manga":
				db.RemoveFromMangaCache(updCacheData.Name)
			case "ranobe":
				db.RemoveFromRanobeCache(updCacheData.Name)
			default:
			}
		case 5:
			loadTemplates()
		default:
		}
	}

}
