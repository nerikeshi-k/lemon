package query

/* サーバーからクライアントに送る系 */

// RequestSdpPack "requestSdp"
type RequestSdpPack struct {
	PartnerID int `json:"partner_id"`
}

// SdpPack "responseSdp" "responseAnswerSdp"
type SdpPack struct {
	EncodedSDP string `json:"encoded_sdp"`
}

// SdpToPartnerPack "requestAnswerSdp" "setAnswerSdp"
type SdpToPartnerPack struct {
	PartnerID  int    `json:"partner_id"`
	EncodedSDP string `json:"encoded_sdp"`
}

// ActionError error
type ActionError struct {
	message string
}

func (a *ActionError) Error() string {
	return a.message
}

/* クライアントから送られてきたのをサーバーが受け取る系 */

// RequestReconnectData requestReconnectのとき送られてくるdata
type RequestReconnectData struct {
	PartnerID int `json:"partner_id"`
}
