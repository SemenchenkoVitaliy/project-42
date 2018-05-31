package main

import (
	"fmt"
	"github.com/SemenchenkoVitaliy/project-42/servers/apiserver"
	"github.com/SemenchenkoVitaliy/project-42/servers/fileserver"
	"github.com/SemenchenkoVitaliy/project-42/servers/httpserver"
	"github.com/SemenchenkoVitaliy/project-42/servers/lbserver"
	"github.com/SemenchenkoVitaliy/project-42/utils"
)

func main() {
	utils.LoadConfig("./configs")
	switch utils.ServerType {
	case "api":
		apiserver.Start()
	case "file":
		fileserver.Start()
	case "http":
		httpserver.Start()
	case "lb":
		lbserver.Start()
	default:
		utils.LogCritical(fmt.Errorf("Incorrect server type: "+utils.ServerType), "Launch server")
	}
}
