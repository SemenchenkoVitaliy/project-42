// Package common provides config options and error logging functions
package common

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Config contains options provided by config file
// or by command line options
var Config struct {
	HostIP   string
	HostPort int

	Tcp struct {
		HostIP     string
		HostPort   int
		BufferSize uint32
	}

	Db struct {
		HostIP   string
		HostPort int
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

// init reads data provided by json-based config file, parses it to Config
// variable and rewrites config options with command line options if any
func init() {
	configFile, err := ioutil.ReadFile("./config.json")
	if err != nil {
		LogCritical(fmt.Errorf("No config file was supplied"), "read config file")
	}

	json.Unmarshal(configFile, &Config)
	if err != nil {
		LogCritical(fmt.Errorf("Wrong config file format"), "unmarshal config file")
	}

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

// Log writes error to log file and displays short error to display or writes
// full error directly to screen if error happend when writing data to file
//
// It accepts error and short explanation message which will be displayed on
// screen as well as written to log file
func Log(err error, text string) {
	fmt.Println("\x1B[31mError occured when trying to: " + text + "\x1B[0m")
	writeLog(text + "\n" + err.Error())
}

// Log writes error to log file and displays short error to display or writes
// full error directly to screen if error happend when writing data to file.
// After that it exits programm with error code
//
// It accepts error and short explanation message which will be displayed on
// screen as well as written to log file
func LogCritical(err error, text string) {
	fmt.Println("\x1B[31mError occured when trying to: " + text + "\x1B[0m")
	writeLog(text + "\n" + err.Error())
	os.Exit(1)
}

// writeLog creates log direcotry if it is not exists and writes data provided
// by Log or LogCritical to file
//
// It accepts string which will be written to log file
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
