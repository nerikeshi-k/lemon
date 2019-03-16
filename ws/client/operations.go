package client

// room内の2人以上のユーザーがやりとりするもの

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/nerikeshi-k/lemon/config"
	"github.com/nerikeshi-k/lemon/ws/client/tradelock"
	"github.com/nerikeshi-k/lemon/ws/query"
)

// JoinRoomOperation ルームに入ってきたと、すでにいる人全員との間でSDPを交換させる
func (room *Room) JoinRoomOperation(newComerID int) error {
	allMemberIDList := room.GetMemberIDList()
	for _, id := range allMemberIDList {
		if id != newComerID {
			go room.TradeSdpOperation(newComerID, id)
		}
	}
	return nil
}

// ReconnectOperation clientからの依頼で対象partnerとの再接続を狙う
func (room *Room) ReconnectOperation(client Member, q query.Query) error {
	data := query.RequestReconnectData{}
	err := json.Unmarshal([]byte(q.Data), &data)
	if err != nil {
		errq := query.CreateError("parse error", q.ConversationKey)
		client.Send(errq.ToJSON())
		return err
	}

	_, ok := room.GetMember(data.PartnerID)
	if !ok {
		errq := query.CreateError("partner does not exist", q.ConversationKey)
		client.Send(errq.ToJSON())
		return errors.New("partner does not exist")
	}

	resq := query.CreateACK(q.ConversationKey)
	client.Send(resq.ToJSON())

	go room.TradeSdpOperation(client.id, data.PartnerID)

	return nil
}

// TradeSdpOperation clientとpartnerの間でのSDP交換を行う
func (room *Room) TradeSdpOperation(clientID int, partnerID int) error {
	err := tradelock.Lock(clientID, partnerID)
	if err != nil {
		// すでにtradeが始まっていた場合
		return err
	}
	defer tradelock.Unlock(clientID, partnerID)
	if config.Debug {
		log.Printf("start: trade between %d and %d\n", clientID, partnerID)
	}

	if partnerID == clientID {
		return errors.New("same person")
	}

	client, ok := room.GetMember(clientID)
	if !ok {
		return errors.New("client does not exist")
	}

	partner, ok := room.GetMember(partnerID)
	if !ok {
		return errors.New("partner does not exist")
	}

	if config.Debug {
		log.Printf("wait for client sdp: trade between %d and %d\n", clientID, partnerID)
	}
	// clientにlocalSdpを作ってもらい送ってもらう
	clientReply, err := client.RequestSdpAction(partner.id)
	if err != nil {
		return err
	}

	if config.Debug {
		log.Printf("wait for partner answer sdp: trade between %d and %d\n", clientID, partnerID)
	}
	// partnerにそのlocalSdpに対するanswerSdpを返してもらう
	clientLocalSdpPack := query.SdpPack{}
	err = json.Unmarshal([]byte(clientReply.Data), &clientLocalSdpPack)
	if err != nil {
		return err
	}
	partnerReply, err := partner.RequestAnswerSdpAction(client.id, clientLocalSdpPack.EncodedSDP)
	if err != nil {
		return err
	}

	if config.Debug {
		log.Printf("send client answer sdp: trade between %d and %d\n", clientID, partnerID)
	}
	// partnerからもらったanswerSdpをclientに返す
	partnerAnswerSdpPack := query.SdpToPartnerPack{}
	err = json.Unmarshal([]byte(partnerReply.Data), &partnerAnswerSdpPack)
	if err != nil {
		return &query.ActionError{}
	}
	client.SetAnswerSdpAction(partner.id, partnerAnswerSdpPack.EncodedSDP)

	if config.Debug {
		log.Printf("finished: trade between %d and %d\n", clientID, partnerID)
	}
	return nil
}
