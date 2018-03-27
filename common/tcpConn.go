package common

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
)

func WriteToStream(conn net.Conn, reqType uint8, reqData []byte) error {
	var reqDataLen uint32 = uint32(len(reqData))
	reqLen := reqDataLen

	if reqDataLen%Config.BufferSize != 0 {
		reqLen += Config.BufferSize - (reqDataLen % Config.BufferSize)
	}

	buffer := new(bytes.Buffer)
	err := binary.Write(buffer, binary.LittleEndian, reqDataLen)
	if err != nil {
		return err
	}
	data := []byte{}
	data = append(data, buffer.Bytes()...)
	data = append(data, byte(reqType))
	data = append(data, reqData...)

	dataToWrite := make([]byte, reqLen)
	copy(dataToWrite, data)

	conn.Write(dataToWrite)
	return nil
}

func ReadFromStream(conn net.Conn) (uint8, string, error) {
	result := []byte{}
	buf := make([]byte, Config.BufferSize)

	n, err := conn.Read(buf)
	if err != nil {
		return 0, "", err
	}

	var temp uint32
	b := buf[:4]
	buffer := bytes.NewReader(b)
	err = binary.Read(buffer, binary.LittleEndian, &temp)
	if err != nil {
		return 0, "", err
	}

	reqSize := temp
	reqType := uint8(buf[4])
	result = append(result, buf[5:n]...)

	curLen := n - 5
	for curLen < int(reqSize) {
		n, err = conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return 0, "", err
			}
		}
		result = append(result, buf[:n]...)
		curLen += n
	}

	reqData := string(result[:reqSize])

	return reqType, reqData, nil
}
