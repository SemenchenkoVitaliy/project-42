package fileServer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/SemenchenkoVitaliy/project-42/common"
)

type urlFile struct {
	Path string
	Url  string
}

type dataFile struct {
	Path string
	Data []byte
}

func processData(reqType uint8, reqData string) (uint8, string) {
	switch reqType {
	case 0:
		dir := fmt.Sprintf("%v/%v", common.Config.SrcDir, reqData)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.Mkdir(dir, 0777)
			if err != nil {
				common.CreateLog(err, fmt.Sprintf("create directory: %v", dir))
				return 1, err.Error()
			}
		}
	case 1:
		data := []urlFile{}
		err := json.Unmarshal([]byte(reqData), &data)
		if err != nil {
			common.CreateLog(err, "parse JSON to load files")
			return 1, err.Error()
		}

		for _, item := range data {
			resp, err := http.Get(item.Url)
			if err != nil {
				common.CreateLog(err, fmt.Sprintf("get http page: %v", item.Url))
				return 1, err.Error()
			}

			bytes, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				common.CreateLog(err, fmt.Sprintf("convert http page to byte slice: %v", item.Url))
				return 1, err.Error()
			}

			name := item.Url[strings.LastIndex(item.Url, "/"):]

			err = ioutil.WriteFile(common.Config.SrcDir+"/"+item.Path+name, bytes, 0777)
			if err != nil {
				common.CreateLog(err, fmt.Sprintf("write file: %v", common.Config.SrcDir+item.Path+name))
				return 1, err.Error()
			}
		}
	case 2:
		data := []dataFile{}
		err := json.Unmarshal([]byte(reqData), &data)
		if err != nil {
			common.CreateLog(err, "parse JSON to write files")
			return 1, err.Error()
		}

		for _, item := range data {
			err = ioutil.WriteFile(common.Config.SrcDir+"/"+item.Path, item.Data, 0777)
			if err != nil {
				common.CreateLog(err, fmt.Sprintf("write file: %v", common.Config.SrcDir+item.Path))
				return 1, err.Error()
			}
		}
	default:
		return 1, "server does not support such command"
	}

	return 0, "ok"
}

func handleRequest(conn net.Conn) {
	fmt.Println(conn.RemoteAddr().String() + " connected")
	for {
		reqType, reqData, err := common.ReadFromStream(conn)
		if err != nil {
			if err == io.EOF {
				break
			}
			common.CreateLog(err, fmt.Sprintf("read from connection stream"))
		}

		respType, respData := processData(reqType, reqData)

		err = common.WriteToStream(conn, respType, []byte(respData))
		if err != nil {
			common.CreateLog(err, fmt.Sprintf("write to connection stream"))
		}
	}
	fmt.Println(conn.RemoteAddr().String() + " disconnected")
}

func openTCPServer() {
	listener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", common.Config.Tcp.Host, common.Config.Tcp.Port))
	if err != nil {
		common.CreateLog(err, fmt.Sprintf("open TCP server on %v:%v", common.Config.Tcp.Host, common.Config.Tcp.Port))
	}
	defer listener.Close()

	fmt.Printf("TCP server is opened on %v:%v\n", common.Config.Tcp.Host, common.Config.Tcp.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			common.CreateLog(err, fmt.Sprintf("accept connection TCP server listener"))
			continue
		}
		go handleRequest(conn)
	}
}
