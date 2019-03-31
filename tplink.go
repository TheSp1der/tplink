package tplink

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"time"
)

// SysInfo A type used to return information from tplink devices
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
			RelayState bool   `json:"relay_state"`
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

// Tplink Device host indentification
type Tplink struct {
	Host string
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

func send(host string, dataSend []byte) ([]byte, error) {
	var header = make([]byte, 4)

	// establish connection (two second timeout)
	conn, err := net.DialTimeout("tcp", host+":9999", time.Duration(time.Second*2))
	if err != nil {
		return []byte(""), err
	}
	defer conn.Close()

	// submit data to device
	writer := bufio.NewWriter(conn)
	_, err = writer.Write(dataSend)
	if err != nil {
		return []byte(""), err
	}
	writer.Flush()

	// read response header to determine response size
	headerReader := io.LimitReader(conn, int64(4))
	_, err = headerReader.Read(header)
	if err != nil {
		return []byte(""), err
	}

	// read response
	respSize := int64(binary.BigEndian.Uint32(header))
	respReader := io.LimitReader(conn, respSize)
	var response = make([]byte, respSize)
	_, err = respReader.Read(response)
	if err != nil {
		return []byte(""), err
	}

	return response, nil
}

// SystemInfo Returns information from targeted device
func (s *Tplink) SystemInfo() (SysInfo, error) {
	var (
		payload  getSysInfo
		jsonResp SysInfo
	)

	j, _ := json.Marshal(payload)

	data := encrypt(string(j))
	resp, err := send(s.Host, data)
	if err != nil {
		return jsonResp, err
	}

	if err := json.Unmarshal([]byte(decrypt(resp)), &jsonResp); err != nil {
		return jsonResp, err
	}
	return jsonResp, nil
}

// TurnOn Device state change to turn remote device on
func (s *Tplink) TurnOn() error {
	var payload changeState

	payload.System.SetRelayState.State = 1

	j, _ := json.Marshal(payload)
	data := encrypt(string(j))
	if _, err := send(s.Host, data); err != nil {
		return err
	}
	return nil
}

// TurnOff Device state change to turn remote device off
func (s *Tplink) TurnOff() error {
	var payload changeState

	payload.System.SetRelayState.State = 0

	j, _ := json.Marshal(payload)
	data := encrypt(string(j))
	if _, err := send(s.Host, data); err != nil {
		return err
	}
	return nil
}
