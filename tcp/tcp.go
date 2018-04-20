// Package tcp contains tools for working with tcp connection
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

// AuthData stores data about worker server which will be sent to main load
// balancing server
type AuthData struct {
	IP   string
	Port int
	Type string
}

// WriteFileData stores data about file which will be sent to file server and
// written to disk
type WriteFileData struct {
	Path string
	Data []byte
}

// Update Cache stores data about product which was removed or modified and
// cache about it should be refreshed
type UpdateCache struct {
	Product  string
	Name     string
	Chapter  int
	Pages    bool
	PagesAll bool
}

// Server stores data about server tcp connection
type Server struct {
	conn      net.Conn
	chInData  chan []byte
	chInType  chan uint8
	chOutData chan []byte
	chOutType chan uint8
	chQuit    chan bool
}

// Start initializes Server struct and enables data recieving and sending
//
// It accepts tcp connection and data handler function
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

// Auth sends authentification data to main server
func (server *Server) Auth(data AuthData) (err error) {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return server.Send(b, 1)
}

// WriteFile sends file info to main server which will redirect it to file
// servers
func (server *Server) WriteFile(path string, fileData []byte) {
	data := WriteFileData{
		Path: path,
		Data: fileData,
	}
	b, _ := json.Marshal(data)
	server.Send(b, 2)
}

// MkDir sends directory info to main server which will redirect it to file
// servers
func (server *Server) MkDir(path string) {
	server.Send([]byte(path), 3)
}

// Recieve reads data from tcp connection
//
// It returns data itself which was read from tcp connection, its type which
// should be used to parse data correctly and error if any
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

// Send writes data to tcp connection
//
// It accepts data and its type by which it should be parsed later
func (server *Server) Send(data []byte, dataType uint8) (err error) {
	server.chOutType <- dataType
	server.chOutData <- data
	return err
}

// Info returns address of local tcp server
func (server Server) Info() (addr string) {
	return server.conn.LocalAddr().String()
}

// Info returns address of remote tcp server
func (server Server) RemoteInfo() (addr string) {
	return server.conn.RemoteAddr().String()
}

// Disconnect forces tcp connection to close
func (server Server) Disconnect() {
	server.chQuit <- true
}

// createReadChan reads data from tcp connection and writes it to channels
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

// createWriteChan reads data from channels and writes it to tcp connection
func (server Server) createWriteChan() {
	for {
		dataType, ok := <-server.chOutType
		if !ok {
			<-server.chOutData
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
