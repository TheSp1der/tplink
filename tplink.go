package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
)

type tplink struct {
	HostName string
}

type getSysInfo struct {
	System struct {
		SysInfo struct {
		} `json:"get_sysinfo"`
	} `json:"system"`
}

type changeState struct {
	System struct {
		SetRelayState struct {
			State int `json:"state"`
		} `json:"set_relay_state"`
	} `json:"system"`
}

func encrypt(plaintext string) []byte {
	n := len(plaintext)
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint32(n))
	ciphertext := []byte(buf.Bytes())

	key := byte(0xAB)
	payload := make([]byte, n)
	for i := 0; i < n; i++ {
		payload[i] = plaintext[i] ^ key
		key = payload[i]
	}

	for i := 0; i < len(payload); i++ {
		ciphertext = append(ciphertext, payload[i])
	}

	return ciphertext
}

func decrypt(ciphertext []byte) string {
	n := len(ciphertext)
	key := byte(0xAB)
	var nextKey byte
	for i := 0; i < n; i++ {
		nextKey = ciphertext[i]
		ciphertext[i] = ciphertext[i] ^ key
		key = nextKey
	}
	return string(ciphertext)
}

func send(hostname string, output []byte) (data []byte, err error) {
	conn, err := net.Dial("tcp", hostname+":9999")
	if err != nil {
		return []byte(""), err
	}
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	_, err = writer.Write(output)
	if err != nil {
		return []byte(""), err
	}
	writer.Flush()

	response, err := readHeader(conn)
	if err != nil {
		return []byte(""), err
	}

	payload, err := readPayload(conn, payloadLength(response))
	if err != nil {
		return []byte(""), err
	}

	return payload, nil
}

func readHeader(conn net.Conn) ([]byte, error) {
	headerReader := io.LimitReader(conn, int64(4))
	var response = make([]byte, 4)
	_, err := headerReader.Read(response)
	return response, err
}

func readPayload(conn net.Conn, length uint32) ([]byte, error) {
	payloadReader := io.LimitReader(conn, int64(length))
	var payload = make([]byte, length)
	_, err := payloadReader.Read(payload)
	return payload, err
}

func payloadLength(header []byte) uint32 {
	payloadLength := binary.BigEndian.Uint32(header)
	return payloadLength
}

func (s *tplink) SystemInfo() (results string, err error) {
	var payload getSysInfo

	j, _ := json.Marshal(payload)

	data := encrypt(string(j))
	reading, err := send(s.HostName, data)
	if err == nil {
		results = decrypt(reading)
	}
	return
}

func (s *tplink) TurnOn() (err error) {
	var payload changeState

	payload.System.SetRelayState.State = 1

	j, _ := json.Marshal(payload)
	data := encrypt(string(j))
	_, err = send(s.HostName, data)
	return
}

func (s *tplink) TurnOff() (err error) {
	var payload changeState

	payload.System.SetRelayState.State = 0

	j, _ := json.Marshal(payload)
	data := encrypt(string(j))
	_, err = send(s.HostName, data)
	return
}
