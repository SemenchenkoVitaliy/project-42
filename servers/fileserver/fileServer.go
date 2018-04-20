package fileserver

import (
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/SemenchenkoVitaliy/project-42/common"
	"github.com/SemenchenkoVitaliy/project-42/tcp"
)

type fileCache struct {
	cache map[string][]byte
	sync.Mutex
}

func NewFileCache() *fileCache {
	return &fileCache{cache: make(map[string][]byte)}
}

func (fc *fileCache) Add(path string, data []byte) {
	fc.Lock()
	defer fc.Unlock()

	fc.cache[path] = data
}

func (fc *fileCache) Find(path string) (data []byte, ok bool) {
	fc.Lock()
	defer fc.Unlock()
	data, ok = fc.cache[path]
	return data, ok
}

func (fc *fileCache) Remove(path string) {
	fc.Lock()
	defer fc.Unlock()

	delete(fc.cache, path)
}

func fsHashedHandler(dir string) http.HandlerFunc {
	cache := NewFileCache()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("403 - access forbidden"))
			return
		}
		path := dir + "/" + r.URL.Path
		fd, ok := cache.Find(path)
		if !ok {
			h := sha256.New()
			h.Write([]byte(r.URL.Path))
			data, err := ioutil.ReadFile(dir + "/" + base64.URLEncoding.EncodeToString(h.Sum(nil)))
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("404 - not found"))
				return
			}
			cache.Add(path, data)
			w.Write(data)
		} else {
			w.Write(fd)
		}
	})
}

func fsHandler(dir string) http.HandlerFunc {
	cache := NewFileCache()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("403 - access forbidden"))
			return
		}
		path := dir + "/" + r.URL.Path
		fd, ok := cache.Find(path)
		if !ok {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("404 - not found"))
				return
			}
			cache.Add(path, data)
			w.Write(data)
		} else {
			w.Write(fd)
		}
	})
}

func tcpHandler(server tcp.Server) {
	err := server.Auth(tcp.AuthData{
		IP:   common.Config.HostIP,
		Port: common.Config.HostPort,
		Type: "file",
	})
	if err != nil {
		common.LogCritical(err, "unable to authentifacate")
		return
	}
	for {
		d, dt, e := server.Recieve()
		if e != nil {
			common.LogCritical(err, "unable to recieve a message from server")
			return
		}
		switch dt {
		case 0:
			continue
		case 2:
			var fileData tcp.WriteFileData
			err = json.Unmarshal(d, &fileData)
			if err != nil {
				common.Log(err, "encode file to write")
				continue
			}
			ioutil.WriteFile(common.Config.SrcDir+"/"+fileData.Path, fileData.Data, 0777)
		case 3:
			dir := string(d)
			err = os.Mkdir(dir, 0777)
			if err != nil {
				common.Log(err, "create directory: "+dir)
				continue
			}
		default:
		}
	}
}

func tcpHashedHandler(server tcp.Server) {
	err := server.Auth(tcp.AuthData{
		IP:   common.Config.HostIP,
		Port: common.Config.HostPort,
		Type: "file",
	})
	if err != nil {
		common.LogCritical(err, "unable to authentifacate")
		return
	}
	for {
		d, dt, e := server.Recieve()
		if e != nil {
			common.LogCritical(err, "unable to recieve a message from server")
			return
		}
		switch dt {
		case 0:
			continue
		case 2:
			var fileData tcp.WriteFileData
			err = json.Unmarshal(d, &fileData)
			if err != nil {
				common.Log(err, "encode file to write")
				continue
			}
			h := sha256.New()
			h.Write([]byte(fileData.Path))
			ioutil.WriteFile(common.Config.SrcDir+"/"+base64.URLEncoding.EncodeToString(h.Sum(nil)), fileData.Data, 0777)
		case 3:
			continue
		default:
		}
	}
}

func openHttpServer() {
	switch common.Config.FSType {
	case "normal":
		http.Handle("/", fsHandler(common.Config.SrcDir))
	case "hashed":
		http.Handle("/", fsHashedHandler(common.Config.SrcDir))
	default:
		common.LogCritical(fmt.Errorf("wrong file server type: %v", common.Config.FSType), "choose file server type")
	}

	fmt.Printf("file server is opened on %v:%v\n", common.Config.HostIP, common.Config.HostPort)
	err := http.ListenAndServe(fmt.Sprintf("%v:%v", common.Config.HostIP, common.Config.HostPort), nil)
	if err != nil {
		common.LogCritical(err, fmt.Sprintf("open http server on %v:%v\n", common.Config.HostIP, common.Config.HostPort))
	}

}

func Start() {
	go openHttpServer()

	cert, err := tls.LoadX509KeyPair("certs/cert.pem", "certs/key.pem")
	if err != nil {
		common.LogCritical(err, "load X509 key pair")
	}

	config := tls.Config{Certificates: []tls.Certificate{cert}, InsecureSkipVerify: true}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%v:%v", common.Config.Tcp.HostIP, common.Config.Tcp.HostPort), &config)
	if err != nil {
		common.LogCritical(err, "connect through tcp to main server")
	}

	server := tcp.Server{}
	switch common.Config.FSType {
	case "normal":
		server.Start(conn, tcpHandler)
	case "hashed":
		server.Start(conn, tcpHashedHandler)
	default:
		common.LogCritical(fmt.Errorf("wrong file server type: %v", common.Config.FSType), "choose file server type")
	}
}
