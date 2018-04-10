package lbserver

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/tcp"
)

func openTcpServer() {
	tcpHostname := fmt.Sprintf("%v:%v", common.Config.Tcp.HostIP, common.Config.Tcp.HostPort)

	listener, err := net.Listen("tcp", tcpHostname)
	defer listener.Close()
	if err != nil {
		common.CreateLogCritical(err, "open tcp server on "+tcpHostname)
	}

	fmt.Printf("TCP server is opened on %v\n", tcpHostname)

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go func() {
			fmt.Println(conn.RemoteAddr().String() + " connected")

			var (
				id         int
				workerType string
			)

			server := tcp.Server{}
			server.Start(conn, getHandler(&id, &workerType))

			switch workerType {
			case "file":
				fileServers.Remove(id)
			case "http":
				httpServers.Remove(id)
			case "api":
				apiServers.Remove(id)
			}

			fmt.Println(conn.RemoteAddr().String() + " disconnected")
		}()
	}
}

func getHandler(id *int, workerType *string) tcp.ConnDataHandler {
	return tcp.ConnDataHandler(func(server tcp.Server) {
		authentificated := false
		for {
			d, dt, e := server.Recieve()
			if e != nil {
				return
			}
			switch dt {
			case 0:
				if !authentificated {
					break
				}
			case 1:
				var data tcp.AuthData
				err := json.Unmarshal(d, &data)
				if err != nil {
					server.Disconnect()
					return
				}
				switch data.Type {
				case "file":
					*id = fileServers.Add(server, data)
				case "http":
					*id = httpServers.Add(server, data)
				case "api":
					*id = apiServers.Add(server, data)
				default:
					server.Disconnect()
					return
				}
				*workerType = data.Type
				authentificated = true
				fmt.Printf("%v authentificated as %v server on %v:%v\n", server.RemoteInfo(), data.Type, data.IP, data.Port)
			case 2:
				if !authentificated {
					break
				}
				worker, err := fileServers.GetOne()
				if err != nil {
					continue
				}
				worker.TCPServer.Send(d, 2)
			case 3:
				if !authentificated {
					break
				}
				worker, err := fileServers.GetOne()
				if err != nil {
					continue
				}
				worker.TCPServer.Send(d, 2)
			default:
			}
		}
	})
}
