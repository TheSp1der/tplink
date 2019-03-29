package tplink

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
)

type SysInfo struct {
	System struct {
		GetSysinfo struct {
			SwVer      string `json:"sw_ver"`
			HwVer      string `json:"hw_ver"`
			Type       string `json:"type"`
			Model      string `json:"model"`
			Mac        string `json:"mac"`
			DevName    string `json:"dev_name"`
			Alias      string `json:"alias"`
			RelayState int    `json:"relay_state"`
			OnTime     int    `json:"on_time"`
			ActiveMode string `json:"active_mode"`
			Feature    string `json:"feature"`
			Updating   int    `json:"updating"`
			IconHash   string `json:"icon_hash"`
			Rssi       int    `json:"rssi"`
			LedOff     int    `json:"led_off"`
			LongitudeI int    `json:"longitude_i"`
			LatitudeI  int    `json:"latitude_i"`
			HwID       string `json:"hwId"`
			FwID       string `json:"fwId"`
			DeviceID   string `json:"deviceId"`
			OemID      string `json:"oemId"`
			NextAction struct {
				Type int `json:"type"`
			} `json:"next_action"`
			ErrCode   int     `json:"err_code"`
			MicType   string  `json:"mic_type"`
			Latitude  float64 `json:"latitude"`
			Longitude float64 `json:"longitude"`
		} `json:"get_sysinfo"`
	} `json:"system"`
}

type Tplink struct {
	HostName string
}

type GetSysInfo struct {
	System struct {
		SysInfo struct {
		} `json:"get_sysinfo"`
	} `json:"system"`
}

type ChangeState struct {
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

func (s *Tplink) SystemInfo() (results string, err error) {
	var payload GetSysInfo

	j, _ := json.Marshal(payload)

	data := encrypt(string(j))
	reading, err := send(s.HostName, data)
	if err == nil {
		results = decrypt(reading)
	}
	return
}

func (s *Tplink) TurnOn() (err error) {
	var payload ChangeState

	payload.System.SetRelayState.State = 1

	j, _ := json.Marshal(payload)
	data := encrypt(string(j))
	_, err = send(s.HostName, data)
	return
}

func (s *Tplink) TurnOff() (err error) {
	var payload ChangeState

	payload.System.SetRelayState.State = 0

	j, _ := json.Marshal(payload)
	data := encrypt(string(j))
	_, err = send(s.HostName, data)
	return
}
