package utils

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
)

type TCPConfig struct {
	IP       string
	Port     int
	CertPath string
	KeyPath  string
	BufSize  uint32
}

type DBConfig struct {
	IP       string
	Port     int
	DbName   string
	User     string
	Password string
}

var Config struct {
	DB  DBConfig
	TCP TCPConfig

	IP        string
	Port      int
	ServerId  int
	LogsDir   string
	SrcDir    string
	PublicUrl string
}

var OauthConfig struct {
	Google struct {
		ClientID     string
		ClientSecret string
	}
}

var ServerType = "lb"

func LoadConfig(configPath string) {
	configFile, err := ioutil.ReadFile(configPath + "/config.json")
	if err != nil {
		LogCritical(fmt.Errorf("No config file was supplied"), "read config file")
	}

	json.Unmarshal(configFile, &Config)
	if err != nil {
		LogCritical(fmt.Errorf("Wrong config file format"), "unmarshal config file")
	}

	if ip := os.Getenv("IP"); ip != "" {
		Config.IP = ip
	}
	if port := os.Getenv("PORT"); port != "" {
		Config.Port, _ = strconv.Atoi(port)
	}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n	%v [options]\n\nParameters:\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&Config.TCP.IP, "tcp-ip", Config.TCP.IP, "tcp server ip")
	flag.IntVar(&Config.TCP.Port, "tcp-port", Config.TCP.Port, "tcp server port")
	flag.StringVar(&Config.TCP.CertPath, "tcp-cert", Config.TCP.CertPath, "tcp server cert path")
	flag.StringVar(&Config.TCP.KeyPath, "tcp-key", Config.TCP.KeyPath, "tcp server key path")

	flag.StringVar(&Config.DB.IP, "db-ip", Config.DB.IP, "database server ip")
	flag.IntVar(&Config.DB.Port, "db-port", Config.DB.Port, "database server port")
	flag.StringVar(&Config.DB.DbName, "db-name", Config.DB.DbName, "database name")
	flag.StringVar(&Config.DB.User, "db-user", Config.DB.User, "database username")
	flag.StringVar(&Config.DB.Password, "db-pwd", Config.DB.Password, "database password")

	flag.StringVar(&Config.IP, "ip", Config.IP, "http server ip")
	flag.IntVar(&Config.Port, "port", Config.Port, "http server port")
	flag.IntVar(&Config.ServerId, "id", Config.ServerId, "tcp server id")
	flag.StringVar(&Config.LogsDir, "logs-dir", Config.LogsDir, "logs directory")
	flag.StringVar(&Config.SrcDir, "files-dir", Config.SrcDir, "directory to store files(for file servers only)")
	flag.StringVar(&Config.PublicUrl, "public-url", Config.PublicUrl, "main http server public domain name")

	flag.StringVar(&ServerType, "server-type", ServerType, "server type(lb, http, api, file)")

	flag.Parse()

	configFile, err = ioutil.ReadFile(configPath + "/oauth.json")
	if err != nil {
		LogCritical(fmt.Errorf("No OAuth config file was supplied"), "read OAuth config file")
	}

	json.Unmarshal(configFile, &OauthConfig)
	if err != nil {
		LogCritical(fmt.Errorf("Wrong OAuth config file format"), "unmarshal OAuth config file")
	}

}
