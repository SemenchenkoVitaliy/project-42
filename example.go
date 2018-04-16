package main

import (
	"fmt"
	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/servers/apiserver"
	"github.com/SemenchenkoVitaliy/project-42/servers/fileserver"
	"github.com/SemenchenkoVitaliy/project-42/servers/httpserver"
	"github.com/SemenchenkoVitaliy/project-42/servers/lbserver"
)

func main() {
	switch common.Config.Server {
	case "api":
		apiserver.Start()
	case "file":
		fileserver.Start()
	case "http":
		httpserver.Start()
	case "lb":
		lbserver.Start()
	default:
		common.CreateLogCritical(fmt.Errorf("Incorrect server type: "+common.Config.Server), "Launch server")
	}
}
