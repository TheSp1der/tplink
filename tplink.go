package tplink

// Credit to:
// sausheong - https://github.com/sausheong/hs1xxplug
// jaedle - https://github.com/jaedle/golang-tplink-hs100/blob/master/internal/connector/connector.go

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"encoding/binary"
	"encoding/json"
)

// This is the key by which all bytes sent/received from tp-link hardware are
// obfuscated.
const hashKey byte = 0xAB
const frameHeaderSize = 4
const frameMaxSize = 1 << 30

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
	Emeter struct {
		GetRealtime struct {
			CurrentMa int `json:"current_ma"`
			VoltageMv int `json:"voltage_mv"`
			PowerMw   int `json:"power_mw"`
			TotalWh   int `json:"total_wh"`
			ErrCode   int `json:"err_code"`
		} `json:"get_realtime"`
		GetVgainIgain struct {
			Vgain   int `json:"vgain"`
			Igain   int `json:"igain"`
			ErrCode int `json:"err_code"`
		} `json:"get_vgain_igain"`
		GetDaystat struct {
			DayList []struct {
				Year     int `json:"year"`
				Month    int `json:"month"`
				Day      int `json:"day"`
				EnergyWh int `json:"energy_wh"`
			} `json:"day_list"`
			ErrCode int `json:"err_code"`
		} `json:"get_daystat"`
	} `json:"emeter"`
}

// Tplink Device host identification
type Tplink struct {
	Host     string
	SwitchID int
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

type changeStateMultiSwitch struct {
	Context struct {
		ChildIds []string `json:"child_ids"`
	} `json:"context"`
	System struct {
		SetRelayState struct {
			State int `json:"state"`
		} `json:"set_relay_state"`
	} `json:"system"`
}

type meterInfo struct {
	System struct {
		GetSysinfo struct{} `json:"get_sysinfo"`
	} `json:"system"`
	Emeter struct {
		GetRealtime   struct{} `json:"get_realtime"`
		GetVgainIgain struct{} `json:"get_vgain_igain"`
	} `json:"emeter"`
}

type dailyStats struct {
	Emeter struct {
		GetDaystat struct {
			Month int `json:"month"`
			Year  int `json:"year"`
		} `json:"get_daystat"`
	} `json:"emeter"`
}

func encrypt(data []byte) []byte {
	ciphertext := make([]byte, len(data))

	key := hashKey
	for i, currentByte := range data {
		encByte := key ^ currentByte
		key = encByte
		ciphertext[i] = encByte
	}

	return ciphertext
}

func decrypt(ciphertext []byte) []byte {
	plaintext := make([]byte, len(ciphertext))

	key := hashKey
	for i, currentByte := range ciphertext {
		decByte := key ^ currentByte
		key = currentByte
		plaintext[i] = decByte
	}

	return plaintext
}

func talkToDevice(host string, payload []byte) ([]byte, error) {
	// establish connection (two second timeout)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, 9999), time.Second*2)
	if err != nil {
		return []byte{}, err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Fatalf("[ERROR] tplink.go: Unable to close connection: %v\n", err)
		}
	}()

	// encrypt payload
	encryptedPayload := encrypt(payload)
	frameData := func(data []byte) []byte {
		frameSize := len(data)
		frameBuffer := make([]byte, frameHeaderSize, frameSize+frameHeaderSize)
		binary.BigEndian.PutUint32(frameBuffer, uint32(frameSize))
		return append(frameBuffer, data...)
	}(encryptedPayload)

	// send payload
	if _, err := conn.Write(frameData); err != nil {
		return nil, fmt.Errorf("[ERROR] tplink.go: Unable to send payload: %v", err)
	}

	// read payload
	data, err := func(r io.Reader) ([]byte, error) {
		header := make([]byte, frameHeaderSize)
		if _, err := io.ReadFull(r, header); err != nil {
			return nil, fmt.Errorf("[ERROR] tplink.go: Unable to read payload: %v", err)
		}

		frameSize := binary.BigEndian.Uint32(header)
		if frameSize > frameMaxSize {
			return nil, fmt.Errorf("[ERROR] tplink.go: Returned data exceeds max frame size: 1GB")
		}

		frameBuffer := make([]byte, frameSize)
		_, err := io.ReadFull(r, frameBuffer)
		if err != nil {
			return nil, fmt.Errorf("[ERROR] tplink.go: Unable to read returned data: %v", err)
		}
		return frameBuffer, nil
	}(conn)
	if err != nil {
		return nil, err
	}

	return decrypt(data), nil
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

	j, err := json.Marshal(payload)
	if err != nil {
		return SysInfo{}, fmt.Errorf("[WARN] Unexpected data: %v", err)
	}

	resp, err := talkToDevice(s.Host, j)
	if err != nil {
		return jsonResp, err
	}

	if err := json.Unmarshal(resp, &jsonResp); err != nil {
		return jsonResp, err
	}
	return jsonResp, nil
}

// ChangeState changes the power state of a single port device
// True = on
// False = off
func (s *Tplink) ChangeState(state bool) error {
	var payload changeState

	if state {
		payload.System.SetRelayState.State = 1
	} else {
		payload.System.SetRelayState.State = 0
	}

	j, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("[WARN] Unexpected data: %v", err)
	}

	if _, err := talkToDevice(s.Host, j); err != nil {
		return err
	}
	return nil
}

// ChangeStateMultiSwitch changes the power state of a device on with multiple outlets/switches
// True = on
// False = off
func (s *Tplink) ChangeStateMultiSwitch(state bool) error {
	var payload changeStateMultiSwitch

	devID, err := getDevID(s)
	if err != nil {
		return err
	}

	payload.Context.ChildIds = append(payload.Context.ChildIds, devID+fmt.Sprintf("%02d", s.SwitchID))
	if state {
		payload.System.SetRelayState.State = 1
	} else {
		payload.System.SetRelayState.State = 0
	}

	j, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("[WARN] Unexpected data: %v", err)
	}
	if _, err := talkToDevice(s.Host, j); err != nil {
		return err
	}
	return nil
}

// GetMeterInfo gets the power stats from a device
func (s *Tplink) GetMeterInto() (SysInfo, error) {
	var (
		payload  meterInfo
		jsonResp SysInfo
	)

	j, err := json.Marshal(payload)
	if err != nil {
		return SysInfo{}, fmt.Errorf("[WARN] Unexpected data: %v", err)
	}

	resp, err := talkToDevice(s.Host, j)
	if err != nil {
		return jsonResp, err
	}

	if err := json.Unmarshal(resp, &jsonResp); err != nil {
		return jsonResp, err
	}
	return jsonResp, nil
}

// GetMeterInfo gets the power stats from a device
func (s *Tplink) GetDailyStats(month, year int) (SysInfo, error) {
	var (
		payload  dailyStats
		jsonResp SysInfo
	)

	payload.Emeter.GetDaystat.Month = month
	payload.Emeter.GetDaystat.Year = year

	j, err := json.Marshal(payload)
	if err != nil {
		return SysInfo{}, fmt.Errorf("[WARN] Unexpected data in response: %v", err)
	}

	resp, err := talkToDevice(s.Host, j)
	if err != nil {
		return jsonResp, err
	}

	if err := json.Unmarshal(resp, &jsonResp); err != nil {
		return jsonResp, err
	}
	return jsonResp, nil
}
