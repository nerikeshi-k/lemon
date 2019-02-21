package client

// サーバーとクライアントの間で完結するやりとり

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/nerikeshi-k/lemon/ws/query"
)

// RequestSdpAction クライアントに対してSDPを送ってくれの要望をする
func (member *Member) RequestSdpAction(partnerID int) (query.Query, error) {
	conv, quit := member.NewConversation()
	defer quit()

	data, _ := json.Marshal(query.RequestSdpPack{
		PartnerID: partnerID,
	})
	conv.Sender <- query.CreateQuery("requestSdp", conv.Key, string(data))

	select {
	case result, ok := <-conv.Receiver:
		if ok {
			return result, nil
		}
		return result, errors.New("received closed")
	case <-time.After(time.Second * 10):
		return query.Query{}, errors.New("timeout")
	}
}

// RequestAnswerSdpAction client側が作ったSDPを受け取らせて、それに対する
// answerSdpを作ってもらい、受け取る
func (member *Member) RequestAnswerSdpAction(partnerID int, encodedSdp string) (query.Query, error) {
	conv, quit := member.NewConversation()
	defer quit()

	data, _ := json.Marshal(query.SdpToPartnerPack{
		PartnerID:  partnerID,
		EncodedSDP: encodedSdp,
	})
	q := query.CreateQuery("requestAnswerSdp", conv.Key, string(data))
	conv.Sender <- q

	select {
	case result, ok := <-conv.Receiver:
		if ok {
			return result, nil
		}
		return result, errors.New("received closed")
	case <-time.After(time.Second * 10):
		return query.Query{}, errors.New("timeout")
	}
}

// SetAnswerSdpAction answerSdpを受け取ってもらう
func (member *Member) SetAnswerSdpAction(partnerID int, encodedSdp string) {
	conv, quit := member.NewConversation()
	defer quit()

	sdpPack := query.SdpToPartnerPack{
		PartnerID:  partnerID,
		EncodedSDP: encodedSdp,
	}
	data, _ := json.Marshal(sdpPack)
	q := query.CreateQuery("setAnswerSdp", conv.Key, string(data))
	conv.Sender <- q
}
