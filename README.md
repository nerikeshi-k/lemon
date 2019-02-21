# lemon
WebRTC Signaling server for chapko

## webhook
未実装です

## websocket

### 接続
`Sec-WebSocket-Protocol` ヘッダにSessionIDを渡してもらう。  
もらったSessionIDを使ってappサーバーに認証をかけ、OKだったら接続する。

#### なぜ `Sec-WebSocket-Protocol` ?
JavaScriptの `new WebSocket()` で他のヘッダを変更できないため。  
仕様的に非常によくなさそうだけど簡単のため一旦はこれでやる。　　
もしも今後ブラウザ以外のクライアントとやりとりするようになったとしたら変えるかもしれない

### 接続完了後のクライアントとのやりとり
JSONで会話する。大まかなフォーマットは以下。  
server -> client も client -> server も同じ
```
{
  "action": string,　// メソッド名みたいなもん
  "conversation_key": string | null, // このkeyで返事してね番号。UUIDとかで適当に作った文字列
  "data": string // JSON形式の文字列. actionによって型が変わるので
}
```
`data` がJSONをstringifyしたもの。JSONの中にJSON……。  
golang側でのパースの簡単さを重視したものだけど、パース問題が解決したらこういうややこしい形はやめるかもしれない。

### サーバから飛んでくるAction

#### "requestSdp"
* clientにlocalSdpを作ってもらう。

from server
```
{
  "action": "requestSdp",
  "conversation_key": string,
  "data" : string
}

data -> {
  "partner_id": number
}
```

from client response
```
{
  "action": "responseSdp", // 実はまあなんでもいい……見てない
  "conversation_key": ↑でもらったkeyを返す
  "data": string
}

data -> {
  "encoded_sdp": string
}
```

encoded_sdpのエンコード方法については深く認知しない。
lemonはSDPを仲介するだけで保持しない。

#### "requestAnswerSdp"
* clientから受け取ったlocalSdpをpartnerに受け取ってもらって、Answerを返してもらう。

from server
```
{
  "action": "requestAnswerSdp",
  "conversation_key": string,
  "data": string
}

data -> {
  "partner_id": number,
  "encoded_sdp": string
}
```

from partner response
```
{
  "action": "responseAnswerSdp", // これも実はまあなんでもいい
  "conversation_key": ↑でもらったkeyを返す
  "data": string
}

data -> {
  "encoded_sdp": string
}
```

#### "setAnswerSdp"
* "requestAnswerSdp" に返答したclientのanswerSdpをpartnerにわたす

from server
```
{
  "action": "setAnswerSdp",
  "conversation_key": string,
  "data" : string
}

data -> {
  "partner_id": number,
  "encoded_sdp": string
}
```

返答は不要。
"requestAnswerSdp" が終わったあとに飛んでくるので、IDを照合してクライアント側でくっつけること