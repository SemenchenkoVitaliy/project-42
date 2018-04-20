package apiserver

import (
	"encoding/json"

	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/mangaLoader"
	dbDriver "github.com/SemenchenkoVitaliy/project-42/mongoDriver"
	"github.com/SemenchenkoVitaliy/project-42/tcp"
)

var mainServer tcp.Server

func tcpHandler(server tcp.Server) {
	mainServer = server
	mangaLoader.Init(&server)

	err := server.Auth(tcp.AuthData{
		IP:   common.Config.HostIP,
		Port: common.Config.HostPort,
		Type: "api",
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

func UpdateProductCache(Product, Name string) {
	data := tcp.UpdateCache{
		Product:  Product,
		Name:     Name,
		Chapter:  0,
		Pages:    false,
		PagesAll: false,
	}
	b, _ := json.Marshal(data)
	mainServer.Send(b, 4)
}

func UpdateProductPagesAllCache(Product, Name string) {
	data := tcp.UpdateCache{
		Product:  Product,
		Name:     Name,
		Chapter:  0,
		Pages:    true,
		PagesAll: true,
	}
	b, _ := json.Marshal(data)
	mainServer.Send(b, 4)
}

func UpdateProductPagesCache(Product, Name string, Chapter int) {
	data := tcp.UpdateCache{
		Product:  Product,
		Name:     Name,
		Chapter:  Chapter,
		Pages:    true,
		PagesAll: false,
	}
	b, _ := json.Marshal(data)
	mainServer.Send(b, 4)
}
