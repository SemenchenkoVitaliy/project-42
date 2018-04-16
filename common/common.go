package common

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

type serverConn struct {
	HostIP   string
	HostPort int
}

var Config struct {
	serverConn

	Tcp struct {
		serverConn
		BufferSize uint32
	}

	Db struct {
		serverConn

		DbName   string
		User     string
		Password string
	}

	PublicUrl string
	LogsDir   string
	Server    string
	SrcDir    string
	FSType    string
}

func init() {
	configFile, err := ioutil.ReadFile("./config.json")
	if err != nil {
		LogCritical(fmt.Errorf("No config file was supplied"), "read config file")
	}

	json.Unmarshal(configFile, &Config)
	if err != nil {
		LogCritical(fmt.Errorf("Wrong config file format"), "unmarshal config file")
	}
	getCmdLineOptions()
}

func getCmdLineOptions() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n	%v[options]\n\nParameters:\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.StringVar(&Config.HostIP, "host", Config.HostIP, "host to open http server")
	flag.IntVar(&Config.HostPort, "port", Config.HostPort, "port to open http server")

	flag.StringVar(&Config.Tcp.HostIP, "tcp-host", Config.Tcp.HostIP, "main server tcp host")
	flag.IntVar(&Config.Tcp.HostPort, "tcp-port", Config.Tcp.HostPort, "main server tcp port")

	flag.StringVar(&Config.Db.HostIP, "db-host", Config.Db.HostIP, "host to connect to database")
	flag.IntVar(&Config.Db.HostPort, "db-port", Config.Db.HostPort, "port to connect to database")

	flag.StringVar(&Config.Db.DbName, "db-name", Config.Db.DbName, "database name")
	flag.StringVar(&Config.Db.User, "db-user", Config.Db.User, "database username")
	flag.StringVar(&Config.Db.Password, "db-pwd", Config.Db.Password, "database password")

	flag.StringVar(&Config.Server, "server-type", Config.Server, "server type(lb, http, api, file)")
	flag.StringVar(&Config.LogsDir, "logs-dir", Config.LogsDir, "logs directory")
	flag.StringVar(&Config.SrcDir, "files-dir", Config.SrcDir, "directory to store files(for file servers only)")

	flag.Parse()
}

func Log(err error, text string) {
	fmt.Println("\x1B[31mError occured when trying to: " + text + "\x1B[0m")
	writeLog(text + "\n" + err.Error())
}

func LogCritical(err error, text string) {
	fmt.Println("\x1B[31mError occured when trying to: " + text + "\x1B[0m")
	writeLog(text + "\n" + err.Error())
	os.Exit(1)
}

func writeLog(text string) {
	if _, err := os.Stat(Config.LogsDir); os.IsNotExist(err) {
		err = os.Mkdir(Config.LogsDir, 0777)
		if err != nil {
			fmt.Printf("\x1B[31mError occured when trying to create log directory: %v\n Error text: %v\n\x1B[0m", err.Error(), text)
			return
		}
	}

	fName := fmt.Sprintf("%v/%v.log", Config.LogsDir, time.Now().Format("2006-01-02-15:04:05"))
	err := ioutil.WriteFile(fName, []byte(text), 0777)
	if err != nil {
		fmt.Printf("\x1B[31mError occured when trying to write log file: %v\n Error text: %v\n\x1B[0m", err.Error(), text)
	}
}
