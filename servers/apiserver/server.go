package apiserver

import (
	"github.com/SemenchenkoVitaliy/project-42/netutils"
	"github.com/SemenchenkoVitaliy/project-42/utils"
)

func Start() {
	fsPublicUrl = utils.Config.PublicUrl
	publicUrl = utils.Config.PublicUrl

	db.Connect(utils.Config.DB.User,
		utils.Config.DB.Password,
		utils.Config.DB.DbName,
		utils.Config.DB.IP,
		utils.Config.DB.Port)
	dbUsers.Connect(utils.Config.DB.User,
		utils.Config.DB.Password,
		utils.Config.DB.DbName,
		utils.Config.DB.IP,
		utils.Config.DB.Port)

	sHTTP := netutils.NewHTTPServer()

	sHTTP.AddRoute("/", "GET", indexGET)
	sHTTP.AddRoute("/", "POST", indexPOST)

	sHTTP.AddRoute("/manga", "GET", mangaGET)
	sHTTP.AddRoute("/manga", "POST", mangaPOST)

	sHTTP.AddRoute("/manga/{name}", "GET", mangaInfoGET)
	sHTTP.AddRoute("/manga/{name}", "POST", mangaInfoPOST)

	sHTTP.AddRoute("/manga/{name}/{chapter}", "GET", mangaChapterGET)
	sHTTP.AddRoute("/manga/{name}/{chapter}", "POST", mangaChapterPOST)

	go sHTTP.Listen(utils.Config.IP, utils.Config.Port)

	cTCP := netutils.NewTCPClient()

	cTCP.Cert(utils.Config.TCP.CertPath, utils.Config.TCP.KeyPath)
	cTCP.Connect(utils.Config.TCP.IP, utils.Config.TCP.Port, tcpHandler)

}
