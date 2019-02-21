package query

import "encoding/json"

// Query クライアントとサーバーとが会話するときに使うJSON
type Query struct {
	Action          string `json:"action"`
	ConversationKey string `json:"conversation_key"`
	Data            string `json:"data"`
}

type errorPack struct {
	Message string `json:"message"`
}

// ToJSON JSON stringにする
func (q *Query) ToJSON() []byte {
	result, _ := json.Marshal(q)
	return result
}

// ParseMessage JSON形式で送られてくるはずのクエリをパースしてQueryにして返す
func ParseMessage(message []byte) (Query, error) {
	query := Query{}
	err := json.Unmarshal(message, &query)
	return query, err
}

// CreateQuery Query構造体を作って返す
func CreateQuery(action string, conversationKey string, data string) Query {
	query := Query{
		Action:          action,
		ConversationKey: conversationKey,
		Data:            data,
	}
	return query
}

// CreateError Error時のQueryを作って返す
func CreateError(message string, conversationKey string) Query {
	data, _ := json.Marshal(errorPack{message})
	return Query{
		Action:          "error",
		ConversationKey: conversationKey,
		Data:            string(data),
	}
}

// CreateACK ack
func CreateACK(conversationKey string) Query {
	return CreateQuery("ack", conversationKey, "")
}
