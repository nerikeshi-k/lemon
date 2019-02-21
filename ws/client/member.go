package client

// キーに合致するConversationに受け取ったクエリを何でもかんでも振り分ける

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nerikeshi-k/lemon/ws/client/ws"
	"github.com/nerikeshi-k/lemon/ws/query"
)

// Member 1メンバーとやりとりするクライアント
type Member struct {
	id            int
	room          *Room
	ws            *ws.Client
	conversations map[string](Conversation)
	mutex         *sync.Mutex
}

// NewMember Mlientのコンストラクタ
func NewMember(id int, room *Room, conn *websocket.Conn) *Member {
	wsClient := ws.New(conn)
	return &Member{
		id:            id,
		room:          room,
		ws:            wsClient,
		mutex:         &sync.Mutex{},
		conversations: map[string](Conversation){},
	}
}

// Conversation 会話の一個の単位
type Conversation struct {
	Key      string
	Receiver chan query.Query
	Sender   chan query.Query
}

const (
	conversationReceiveBuffer = 16
	conversationSendBuffer    = 16
)

// Start メッセージの送受信と処理を開始
func (m *Member) Start(ctx context.Context) {
	serveStopped := make(chan struct{})

	defer m.room.RemoveMember(m.id)

	go func() {
		defer func() {
			m.ws.Destruct()
		}()

		for {
			select {
			case <-serveStopped:
				return
			case message, ok := <-m.ws.Receiver:
				if !ok {
					return
				}
				q, err := query.ParseMessage(message)
				if err != nil {
					errq := query.CreateError("could not parse message", q.ConversationKey)
					m.ws.Send(errq.ToJSON())
					break
				}
				m.switchReceivedQuery(q)
			}
		}
	}()

	// 部屋にいる人達とSDPの交換をする
	m.room.JoinRoomOperation(m.id)

	m.ws.Serve(ctx, serveStopped)
}

func (m *Member) switchReceivedQuery(q query.Query) {
	key := q.ConversationKey
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// キーがあった場合はそのconversationに入力を振り分ける
	// なかった場合はactionの値に応じてoperationを作る
	if conv, ok := m.conversations[key]; ok {
		if len(conv.Receiver) == conversationReceiveBuffer {
			// バッファがいっぱいになっていた場合は捨てる
			return
		}
		conv.Receiver <- q
	} else {
		m.manageRequestFromClient(q)
	}
}

func (m *Member) manageRequestFromClient(q query.Query) {
	switch q.Action {
	case "requestReconnect":
		m.room.ReconnectOperation(*m, q)
	}
}

// NewConversation 入出力を作る
func (m *Member) NewConversation() (Conversation, func()) {
	key := uuid.New().String()
	conv := Conversation{
		Key:      key,
		Receiver: make(chan query.Query, conversationReceiveBuffer),
		Sender:   make(chan query.Query, conversationSendBuffer),
	}

	m.mutex.Lock()
	m.conversations[key] = conv
	m.mutex.Unlock()

	go func() {
		for s := range conv.Sender {
			message, err := json.Marshal(s)
			if err != nil {
				break
			}
			m.Send(message)
		}
	}()

	quit := func() {
		m.mutex.Lock()
		defer m.mutex.Unlock()

		_, exists := m.conversations[key]
		if !exists {
			return
		}
		delete(m.conversations, key)
		close(conv.Receiver)
		close(conv.Sender)
	}
	return conv, quit
}

// Send メッセージを送る
func (m *Member) Send(message []byte) {
	m.ws.Send(message)
}
