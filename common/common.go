package common

import (
	"encoding/json"
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
			fmt.Println("\x1B[31mError occured when trying to create log directory" + err.Error() + "\x1B[0m")
			fmt.Println("\x1B[31mError text: " + text + "\x1B[0m")
		}
	}

	name := time.Now().Format("2006-01-02_15:04:05_+0000_UTC_m=+0.000000001") + ".log"
	err := ioutil.WriteFile(Config.LogsDir+"/"+name, []byte(text), 0777)
	if err != nil {
		fmt.Println("\x1B[31mError occured when trying to write log file" + err.Error() + "\x1B[0m")
		fmt.Println("\x1B[31mError text: " + text + "\x1B[0m")
	}
}
