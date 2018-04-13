package lbserver

import (
	"fmt"

	"github.com/SemenchenkoVitaliy/project-42/tcp"
)

var (
	fileServers serversPull
	apiServers  serversPull
	httpServers serversPull
)

type workerServer struct {
	ID        int
	TCPServer tcp.Server
	IP        string
	Port      uint32
}

type balanceFunc func() workerServer

type serversPull struct {
	workers map[int]workerServer
	balFunc func() workerServer
	curId   int
}

func (sp *serversPull) getRoundRobin() balanceFunc {
	max := len(sp.workers)
	cur := 0
	keys := make([]int, 0, max)
	for k := range sp.workers {
		keys = append(keys, k)
	}
	return balanceFunc(func() workerServer {
		result := sp.workers[keys[cur]]
		if cur++; cur == max {
			cur = 0
		}
		return result
	})
}

func (sp *serversPull) Init(balType string) error {
	sp.workers = make(map[int]workerServer)
	switch balType {
	case "round-robin":
		sp.balFunc = sp.getRoundRobin()
		return nil
	}
	return fmt.Errorf("not such balancer type")
}

func (sp *serversPull) Add(server tcp.Server, authInfo tcp.AuthData) (id int) {
	id = sp.curId
	sp.curId += 1

	sp.workers[id] = workerServer{
		ID:        id,
		TCPServer: server,
		IP:        authInfo.IP,
		Port:      uint32(authInfo.Port),
	}
	sp.balFunc = sp.getRoundRobin()
	return id
}

func (sp *serversPull) Remove(id int) {
	delete(sp.workers, id)
	sp.balFunc = sp.getRoundRobin()
}

func (sp *serversPull) GetOne() (worker workerServer, err error) {
	if len(sp.workers) == 0 {
		return worker, fmt.Errorf("not found")
	}
	return sp.balFunc(), nil
}

func (sp *serversPull) GetAll() (workers []workerServer, err error) {
	if len(sp.workers) == 0 {
		return workers, fmt.Errorf("not found")
	}
	for _, worker := range sp.workers {
		workers = append(workers, worker)
	}
	return workers, nil
}

func Start() {
	fileServers.Init("round-robin")
	apiServers.Init("round-robin")
	httpServers.Init("round-robin")

	go openHttpServer()
	openTcpServer()
}
