package tcp

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/SemenchenkoVitaliy/project-42/common"
)

type ConnDataHandler func(server Server)

type AuthData struct {
	IP   string
	Port int
	Type string
}

type WriteFileData struct {
	Path string
	Data []byte
}

type UpdateCache struct {
	Product  string
	Name     string
	Chapter  int
	Pages    bool
	PagesAll bool
}

type Server struct {
	conn      net.Conn
	chInData  chan []byte
	chInType  chan uint8
	chOutData chan []byte
	chOutType chan uint8
	chQuit    chan bool
}

func (server *Server) Start(conn net.Conn, handler ConnDataHandler) {
	server.conn = conn
	server.chInData, server.chInType = make(chan []byte), make(chan uint8)
	server.chOutData, server.chOutType = make(chan []byte), make(chan uint8)
	server.chQuit = make(chan bool)

	go server.createReadChan()
	go server.createWriteChan()
	go handler(*server)

	_ = <-server.chQuit

	close(server.chInType)
	close(server.chInData)
	close(server.chOutType)
	close(server.chOutData)
	close(server.chQuit)
	server.conn.Close()
}

func (server *Server) Auth(data AuthData) (err error) {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return server.Send(b, 1)
}

func (server *Server) Recieve() (data []byte, dataType uint8, err error) {
	dataType, ok := <-server.chInType
	if !ok {
		return data, dataType, fmt.Errorf("Connection closed")
	}
	data, ok = <-server.chInData
	if !ok {
		return data, dataType, fmt.Errorf("Connection closed")
	}
	return data, dataType, err
}

func (server *Server) Send(data []byte, dataType uint8) (err error) {
	server.chOutType <- dataType
	server.chOutData <- data
	return err
}

func (server Server) Info() (addr string) {
	return server.conn.LocalAddr().String()
}

func (server Server) RemoteInfo() (addr string) {
	return server.conn.RemoteAddr().String()
}

func (server Server) Disconnect() {
	server.chQuit <- true
}

func (server Server) createReadChan() {
	for {
		buf := make([]byte, common.Config.Tcp.BufferSize)
		n, err := server.conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				server.chQuit <- true
				return
			}
			continue
		}

		dataSize := binary.LittleEndian.Uint32(buf[:4])
		dataType := uint8(buf[4])

		data := make([]byte, 0, dataSize)
		data = append(data, buf[5:n]...)

		readDataSize := n - 5
		for readDataSize < int(dataSize) {
			n, err = server.conn.Read(buf)
			if err != nil {
				if err == io.EOF {
					server.chQuit <- true
					return
				}
				continue
			}
			data = append(data, buf[:n]...)
			readDataSize += n
		}
		server.chInType <- dataType
		server.chInData <- data
	}
}

func (server Server) createWriteChan() {
	for {
		dataType, ok := <-server.chOutType
		if !ok {
			return
		}
		data, ok := <-server.chOutData
		if !ok {
			return
		}
		dataSize := uint32(len(data))
		writeDataSize := dataSize

		if writeDataSize%common.Config.Tcp.BufferSize != 0 {
			writeDataSize += common.Config.Tcp.BufferSize - (writeDataSize % common.Config.Tcp.BufferSize)
		}

		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, dataSize)

		buf := make([]byte, 0, writeDataSize)
		buf = append(buf, b...)
		buf = append(buf, byte(dataType))
		buf = append(buf, data...)

		_, err := server.conn.Write(buf)
		if err != nil {
			server.chQuit <- true
			return
		}

	}
}
