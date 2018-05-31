package netutils

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/SemenchenkoVitaliy/project-42/utils"
)

// ConnDataHandler handles coming and outgoing data based on its type
type ConnDataHandler func(server Server)

// AuthData stores data about worker server which will be sent to main load
// balancing server
type AuthData struct {
	IP   string
	Port int
	Type string
	Id   int
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
	Product string
	Name    string
}

// Server stores data about server tcp connection
type Server struct {
	bufSize   uint32
	conn      net.Conn
	chInData  chan []byte
	chInType  chan uint8
	chOutData chan []byte
	chOutType chan uint8
	chQuit    chan bool
}

// NewServer creates new instance of Server and initializes it with given net.Conn
func NewServer(conn net.Conn, bufferSize uint32) (server *Server) {
	return &Server{
		bufSize: bufferSize,
		conn:    conn,
	}
}

// Start initializes Server struct channels and starts data recieving and sending
//
// It accepts data handler function
func (server *Server) Start(handler ConnDataHandler) {
	server.chInData, server.chInType = make(chan []byte), make(chan uint8)
	server.chOutData, server.chOutType = make(chan []byte), make(chan uint8)
	server.chQuit = make(chan bool)

	defer server.conn.Close()
	defer close(server.chInType)
	defer close(server.chInData)
	defer close(server.chOutType)
	defer close(server.chOutData)
	defer close(server.chQuit)

	go server.createReadChan()
	go server.createWriteChan()
	go handler(*server)

	_ = <-server.chQuit
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

// UpdateHtml sends command to update cached products from database to main
// server which will redirect it to all api and http servers
func (server *Server) UpdateProductCache(Product, Name string) {
	data := UpdateCache{
		Product: Product,
		Name:    Name,
	}
	b, _ := json.Marshal(data)
	server.Send(b, 4)
}

// UpdateHtml sends command to update cached html templates to main server which
// will redirect it to all http servers
func (server *Server) UpdateHtml() {
	server.Send([]byte{}, 5)
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
		buf := make([]byte, server.bufSize)
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

		if writeDataSize%server.bufSize != 0 {
			writeDataSize += server.bufSize - (writeDataSize % server.bufSize)
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

type TCPClient struct {
	cert string
	key  string
}

func NewTCPClient() (s *TCPClient) {
	return &TCPClient{}
}

func (s *TCPClient) Cert(cert, key string) {
	s.cert = cert
	s.key = key
}

func (s *TCPClient) Connect(ip string, port int, tcpHandler ConnDataHandler) {
	cert, err := tls.LoadX509KeyPair(s.cert, s.key)
	if err != nil {
		utils.LogCritical(err, "load X509 key pair")
	}

	config := tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", fmt.Sprintf("%v:%v", ip, port), &config)

	if err != nil {
		utils.LogCritical(err, "connect to tcp server")
	}

	server := NewServer(conn, utils.Config.TCP.BufSize)
	server.Start(tcpHandler)

}

type TCPServer struct {
	cert string
	key  string
}

func NewTCPServer() (s *TCPServer) {
	return &TCPServer{}
}

func (s *TCPServer) Cert(cert, key string) {
	s.cert = cert
	s.key = key
}

func (s *TCPServer) Listen(ip string, port int, tcpHandler ConnDataHandler) {
	hostname := fmt.Sprintf("%v:%v", ip, port)

	cert, err := tls.LoadX509KeyPair(s.cert, s.key)
	if err != nil {
		utils.LogCritical(err, "load X509 key pair")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	listener, err := tls.Listen("tcp", hostname, config)
	if err != nil {
		utils.LogCritical(err, "open tcp server on "+hostname)
	}
	defer listener.Close()

	fmt.Println("TCP server is opened on " + hostname)

	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}
		go NewServer(conn, utils.Config.TCP.BufSize).Start(tcpHandler)
	}
}
