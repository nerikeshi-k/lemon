package apiClient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/nerikeshi-k/lemon/config"
)

// RequestError CodeはHTTP Status Code
type RequestError struct {
	Code    int
	Message []byte
}

func (e *RequestError) Error() string {
	return string(e.Message)
}

// Auth is auth
type Auth struct {
	ID        int    `json:"id"`
	SessionID string `json:"session_id"`
	HashID    string `json:"hash_id"`
	Name      string `json:"name"`
	IsOwner   bool   `json:"is_owner"`
}

// AuthResponse is auth response
type AuthResponse struct {
	Resource Auth `json:"resource"`
}

var client = &http.Client{Timeout: time.Duration(10) * time.Second}

func Authenticate(roomKey string, sessionID string) (*Auth, *RequestError) {
	// 認証チェック
	url := fmt.Sprintf("%s/api/chat/rooms/%s/auth", config.AppServerAddress, roomKey)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-Room-Session-Id", sessionID)
	response, err := client.Do(req)
	if err != nil {
		return nil, &RequestError{
			Code:    400,
			Message: []byte("HTTP Error."),
		}
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, &RequestError{
			Code:    400,
			Message: []byte("Authentication failed."),
		}
	}
	bufbody := new(bytes.Buffer)
	bufbody.ReadFrom(response.Body)
	authResponse := &AuthResponse{}
	err = json.Unmarshal(bufbody.Bytes(), &authResponse)
	if err != nil {
		return nil, &RequestError{
			Code:    500,
			Message: []byte("Parsing Error."),
		}
	}
	return &authResponse.Resource, nil
}

func JoinRoom(roomKey string, sessionID string) *RequestError {
	url := fmt.Sprintf("%s/api/chat/rooms/%s/voicechatgroup/join", config.AppServerAddress, roomKey)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-Room-Session-Id", sessionID)
	response, err := client.Do(req)
	if err != nil {
		return &RequestError{
			Code:    400,
			Message: []byte("HTTP Error."),
		}
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return &RequestError{
			Code:    400,
			Message: []byte("Could not join room."),
		}
	}
	return nil
}

func LeaveRoom(roomKey string, sessionID string) *RequestError {
	url := fmt.Sprintf("%s/api/chat/rooms/%s/voicechatgroup/leave", config.AppServerAddress, roomKey)
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("X-Room-Session-Id", sessionID)
	response, err := client.Do(req)
	if err != nil {
		return &RequestError{
			Code:    400,
			Message: []byte("HTTP Error."),
		}
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return &RequestError{
			Code:    400,
			Message: []byte("Could not leave room."),
		}
	}
	return nil
}
