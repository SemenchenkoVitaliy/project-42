package lbserver

import (
	"encoding/json"
	"fmt"

	"github.com/SemenchenkoVitaliy/project-42/netutils"
	"github.com/SemenchenkoVitaliy/project-42/utils"
)

const fileServersWriteNum int = 3

func tcpHandler(server netutils.Server) {
	fmt.Println(server.RemoteInfo() + " connected")
	defer fmt.Println(server.RemoteInfo() + " disconnected")

	authentificated := false
	for {
		d, dataType, err := server.Recieve()
		if err != nil {
			return
		}
		switch dataType {
		case 0:
			if !authentificated {
				break
			}
		case 1:
			var data netutils.AuthData
			err := json.Unmarshal(d, &data)
			if err != nil {
				server.Disconnect()
				return
			}
			var id int
			switch data.Type {
			case "file":
				if id, ok := fileServers.Add(server, data); ok {
					defer fileServers.Remove(id)
					break
				}
				server.Disconnect()
			case "http":
				id = httpServers.Add(server, data)
				defer httpServers.Remove(id)
			case "api":
				id = apiServers.Add(server, data)
				defer apiServers.Remove(id)
			default:
				server.Disconnect()
				return
			}
			authentificated = true
			fmt.Printf("%v authentificated as %v server on %v:%v\n", server.RemoteInfo(), data.Type, data.IP, data.Port)
		case 2, 3:
			if !authentificated {
				break
			}

			idsAdded := []int{}
			ids, ok := fileServers.GetNIds(fileServersWriteNum)
			if !ok {
				utils.Log(err, "Get file servers ids")
			}
			for _, id := range ids {
				worker, ok := fileServers.GetOne(id)
				if !ok {
					utils.Log(err, fmt.Sprintf("Get file server by id: %v", id))
					continue
				}
				worker.Server.Send(d, dataType)
				idsAdded = append(idsAdded, id)
			}
			var fileData netutils.WriteFileData
			err = json.Unmarshal(d, &fileData)
			if err != nil {
				utils.Log(err, "encode file data to get struct")
				continue
			}

			err = db.AddEntry(fileData.Path, idsAdded)
			if err != nil {
				utils.Log(err, "Add file to balancing db: "+fileData.Path)
			}

			// worker, ok := fileServers.GetOne()
			// if !ok {
			// 	continue
			// }
			// worker.Server.Send(d, dataType)
		case 4:
			if !authentificated {
				break
			}
			if workers, ok := apiServers.GetAll(); ok {
				for _, worker := range workers {
					worker.Server.Send(d, dataType)
				}
			}
			if workers, ok := httpServers.GetAll(); ok {
				for _, worker := range workers {
					worker.Server.Send(d, dataType)
				}
			}
		case 5:
			if !authentificated {
				break
			}
			if workers, ok := httpServers.GetAll(); ok {
				for _, worker := range workers {
					worker.Server.Send(d, dataType)
				}
			}
		default:
		}
	}
}
