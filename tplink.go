package tplink

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"time"

	"encoding/binary"
	"encoding/json"
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
			Children   []struct {
				ID         string `json:"id"`
				State      int    `json:"state"`
				Alias      string `json:"alias"`
				OnTime     int    `json:"on_time"`
				NextAction struct {
					Type int `json:"type"`
				} `json:"next_action"`
			} `json:"children"`
			ChildNum   int `json:"child_num"`
			NtcState   int `json:"ntc_state"`
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

// Tplink Device host identification
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

type changeStateChild struct {
	Context struct {
		ChildIds []string `json:"child_ids"`
	} `json:"context"`
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
	ciphertext := buf.Bytes()

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
		ciphertext[i] ^= key
		key = nextKey
	}
	return string(ciphertext)
}

func send(host string, dataSend []byte) ([]byte, error) {
	var header = make([]byte, 4)

	// establish connection (two second timeout)
	conn, err := net.DialTimeout("tcp", host+":9999", time.Second*2)
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

func getDevID(s *Tplink) (string, error) {
	info, err := s.SystemInfo()
	if err != nil {
		return "", err
	}
	return info.System.GetSysinfo.DeviceID, nil
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

// TurnOnChild Device state change to turn remote device on with multiple controls
func (s *Tplink) TurnOnChild(id int) error {
	var payload changeStateChild

	devID, err := getDevID(s)
	if err != nil {
		return err
	}

	payload.Context.ChildIds = append(payload.Context.ChildIds, devID+fmt.Sprintf("%02d", id))
	payload.System.SetRelayState.State = 1

	j, _ := json.Marshal(payload)
	data := encrypt(string(j))
	if _, err := send(s.Host, data); err != nil {
		return err
	}
	return nil
}

// TurnOffChild Device state change to turn remote device off with multiple controls
func (s *Tplink) TurnOffChild(id int) error {
	var payload changeStateChild

	devID, err := getDevID(s)
	if err != nil {
		return err
	}

	payload.Context.ChildIds = append(payload.Context.ChildIds, devID+fmt.Sprintf("%02d", id))
	payload.System.SetRelayState.State = 0

	j, _ := json.Marshal(payload)
	data := encrypt(string(j))
	if _, err := send(s.Host, data); err != nil {
		return err
	}
	return nil
}
