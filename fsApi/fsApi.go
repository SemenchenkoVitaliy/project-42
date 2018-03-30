package fsApi

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/SemenchenkoVitaliy/project-42/common"
)

var conn net.Conn

type UrlFile struct {
	Path string
	Url  string
}

type DataFile struct {
	Path string
	Data []byte
}

func init() {
	connection, err := net.Dial("tcp", fmt.Sprintf("%v:%v", common.Config.Tcp.Host, common.Config.Tcp.Port))
	if err != nil {
		common.CreateLog(err, fmt.Sprintf("Connect to file server througth TCP %v:%v", common.Config.Tcp.Host, common.Config.Tcp.Port))
		os.Exit(1)
	}
	conn = connection
}

func sendData(reqType uint8, reqData []byte) (uint8, string) {
	err := common.WriteToStream(conn, reqType, reqData)
	if err != nil {
		common.CreateLog(err, fmt.Sprintf("write to stream of file server"))
	}

	respType, respData, err := common.ReadFromStream(conn)
	if err != nil {
		common.CreateLog(err, fmt.Sprintf("read from stream of file server"))
	}

	return respType, respData
}

func MkDir(path string) error {
	respType, respData := sendData(0, []byte(path))
	if respType == 0 {
		return nil
	}
	return fmt.Errorf(respData)
}

func LoadFiles(arr []UrlFile) error {
	data, err := json.Marshal(arr)
	if err != nil {
		return err
	}

	respType, respData := sendData(1, data)
	if respType == 0 {
		return nil
	}
	return fmt.Errorf(respData)
}

func WriteFile(path string, rawData []byte) error {
	data, err := json.Marshal([]DataFile{DataFile{Path: path, Data: rawData}})
	if err != nil {
		return err
	}

	respType, respData := sendData(2, data)
	if respType == 0 {
		return nil
	}
	return fmt.Errorf(respData)
}
