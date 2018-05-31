package httpserver

import (
	"github.com/SemenchenkoVitaliy/project-42/netutils"
	"github.com/SemenchenkoVitaliy/project-42/utils"
)

func Start() {
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

	loadTemplates()
	loadOauthConfig(utils.OauthConfig.Google.ClientID, utils.OauthConfig.Google.ClientSecret)

	sHTTP := netutils.NewHTTPServer()

	sHTTP.AddRoute("/", "GET", index)
	sHTTP.AddRoute("/manga", "GET", manga)
	sHTTP.AddRoute("/manga/{name}", "GET", mangaInfo)
	sHTTP.AddRoute("/manga/{name}/{chapter}", "GET", mangaChapter)

	sHTTP.AddRoute("/admin", "GET", admin)
	sHTTP.AddRoute("/admin/manga/{name}", "GET", adminMangaInfo)
	sHTTP.AddRoute("/login", "GET", login)
	sHTTP.AddRoute("/logout", "GET", logout)
	sHTTP.AddRoute("/googleCallback", "GET", googleCallback)

	sHTTP.AddDir("/static/", "./static/")

	go sHTTP.Listen(utils.Config.IP, utils.Config.Port)

	cTCP := netutils.NewTCPClient()

	cTCP.Cert(utils.Config.TCP.CertPath, utils.Config.TCP.KeyPath)
	cTCP.Connect(utils.Config.TCP.IP, utils.Config.TCP.Port, tcpHandler)
}
