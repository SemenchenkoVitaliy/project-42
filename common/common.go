package common

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"
)

type serverConn struct {
	HostIP   string
	HostPort int
}

type dbConn struct {
	serverConn

	DbName   string
	User     string
	Password string
}

type tcpConn struct {
	serverConn
	BufferSize uint32
}

type Conf struct {
	serverConn

	Tcp tcpConn
	Db  dbConn

	PublicUrl string
	LogsDir   string
	SrcDir    string
}

var Config Conf

func init() {
	configFile, err := os.Open("./config.json")
	defer configFile.Close()
	if err != nil {
		CreateLogCritical(fmt.Errorf("No config file was supplied"), "open config file")
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&Config)
	if err != nil && err != io.EOF {
		CreateLogCritical(fmt.Errorf("Wrong config file format"), "decode config file")
	}
	getCmdLineOptions()
}

func getCmdLineOptions() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n	%v[options]\n\nParameters:\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	host := flag.String("host", Config.HostIP, "host to open http server")
	port := flag.Int("port", Config.HostPort, "port to open http server")

	tcpHost := flag.String("tcp-host", Config.Tcp.HostIP, "main server tcp host")
	tcpPort := flag.Int("tcp-port", Config.Tcp.HostPort, "main server tcp port")

	dbHost := flag.String("db-host", Config.Db.HostIP, "host to connect to database")
	dbPort := flag.Int("db-port", Config.Db.HostPort, "port to connect to database")

	dbName := flag.String("db-name", Config.Db.DbName, "database name")
	dbUser := flag.String("db-user", Config.Db.User, "database username")
	dbPwd := flag.String("db-pwd", Config.Db.Password, "database password")

	logsDir := flag.String("logs-dir", Config.LogsDir, "logs directory")
	srcDir := flag.String("files-dir", Config.SrcDir, "directory to store files(for file servers only)")

	flag.Parse()

	Config.HostIP = *host
	Config.HostPort = *port

	Config.Tcp.HostPort = *tcpPort
	Config.Tcp.HostIP = *tcpHost

	Config.Db.HostIP = *dbHost
	Config.Db.HostPort = *dbPort

	Config.Db.DbName = *dbName
	Config.Db.User = *dbUser
	Config.Db.Password = *dbPwd

	Config.LogsDir = *logsDir
	Config.SrcDir = *srcDir
}

func CreateLog(err error, text string) {
	fmt.Println("\x1B[31mError occured when trying to: " + text + "\x1B[0m")
	writeLog(text + "\n" + err.Error())
}

func CreateLogCritical(err error, text string) {
	fmt.Println("\x1B[31mError occured when trying to: " + text + "\x1B[0m")
	writeLog(text + "\n" + err.Error())
	os.Exit(1)
}

func writeLog(text string) {
	if _, err := os.Stat(Config.LogsDir); os.IsNotExist(err) {
		err = os.Mkdir(Config.LogsDir, 0777)
		if err != nil {
			fmt.Println("\x1B[31mError occured when trying to create log directory: " + err.Error() + "\x1B[0m")
			fmt.Println("\x1B[31mError text: " + text + "\x1B[0m")
			return
		}
	}

	name := time.Now().Format("2006-01-02_15:04:05_+0000_UTC_m=+0.000000001") + ".log"
	err := ioutil.WriteFile(Config.LogsDir+"/"+name, []byte(text), 0777)
	if err != nil {
		fmt.Println("\x1B[31mError occured when trying to write log file: " + err.Error() + "\x1B[0m")
		fmt.Println("\x1B[31mError text: " + text + "\x1B[0m")
	}
}
