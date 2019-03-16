package ws

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/nerikeshi-k/lemon/config"
	"github.com/nerikeshi-k/lemon/ws/apiClient"
	"github.com/nerikeshi-k/lemon/ws/client"
	"github.com/nerikeshi-k/lemon/ws/roomlist"
)

var connectCounter int32

func handler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&connectCounter, 1)
		defer atomic.AddInt32(&connectCounter, -1)

		// URLからroom_keyを受け取る
		query := r.URL.Query()
		roomKey := query.Get("room_key")
		if roomKey == "" {
			w.WriteHeader(400)
			w.Write([]byte("Set room_key on GET query."))
			return
		}

		// SessionIDをヘッダから取得
		// 非常に悪いがブラウザのWebsocketが他にヘッダを送れないので
		sessionID := r.Header.Get("Sec-WebSocket-Protocol")
		if sessionID == "" {
			w.WriteHeader(400)
			w.Write([]byte("Set SessionId in 'Sec-WebSocket-Protocol' Header."))
			return
		}

		// アプリサーバに問い合わせて認証
		auth, reqerr := apiClient.Authenticate(roomKey, sessionID)
		if reqerr != nil {
			w.WriteHeader(reqerr.Code)
			w.Write(reqerr.Message)
			return
		}

		// 認証OKなら入室状態にする
		reqerr = apiClient.JoinRoom(roomKey, sessionID)
		if reqerr != nil {
			w.WriteHeader(reqerr.Code)
			w.Write(reqerr.Message)
			return
		}
		// クライアントとの接続が切れたらLeaveRoomする
		defer func() {
			err := apiClient.LeaveRoom(roomKey, sessionID)
			if err != nil {
				log.Printf("Leave Room failed. room: %s, member id: %d", roomKey, auth.ID)
			}
		}()

		// サブプロトコルを鸚鵡返しする
		subprotocols := make([]string, 1)
		subprotocols = append(subprotocols, sessionID)
		upgrader := websocket.Upgrader{
			CheckOrigin: func(request *http.Request) bool {
				return true
			},
			Subprotocols:   subprotocols,
			ReadBufferSize: 10,
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade error:", err)
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		room := roomlist.GetOrCreateRoom(roomKey)
		client := client.NewMember(auth.ID, &room, ws)
		room.SetMember(*client)
		client.Start(ctx)
	}
}

// Serve メソッド名の通り
func Serve() {
	if config.Debug {
		go func() {
			for {
				time.Sleep(time.Second * 60)
				log.Printf("connecting: %d", connectCounter)
			}
		}()
	}

	rtr := mux.NewRouter()
	rtr.HandleFunc("/", handler())
	hdlr := http.NewServeMux()
	hdlr.Handle("/", rtr)
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", config.BoundAddress, config.WebsocketPort), hdlr)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
