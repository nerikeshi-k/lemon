package tradelock

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
)

type marker struct{}

var mutex *sync.Mutex
var activeTradeList map[string](marker)

func init() {
	mutex = &sync.Mutex{}
	activeTradeList = make(map[string](marker))
}

// Lock ClientとPartnerの間のtrade sdpをロックする。
// すでにlockされていた場合はerrorを返す
func Lock(clientID int, partnerID int) error {
	mutex.Lock()
	defer mutex.Unlock()

	key := generateKey(clientID, partnerID)
	if _, exists := activeTradeList[key]; exists {
		return errors.New("already exists")
	}
	activeTradeList[key] = marker{}
	return nil
}

// Unlock unlockする
func Unlock(clientID int, partnerID int) {
	mutex.Lock()
	defer mutex.Unlock()

	key := generateKey(clientID, partnerID)
	delete(activeTradeList, key)
}

func generateKey(clientID int, partnerID int) string {
	hashedClientID := idToHash(clientID)
	hashedPartnerID := idToHash(partnerID)
	return hex.EncodeToString(integrateHash(hashedClientID, hashedPartnerID))
}

func integrateHash(a []uint8, b []uint8) []uint8 {
	result := make([]uint8, 16)
	for i, v := range a {
		result[i] = b[i] + v
	}
	return result
}

func idToHash(num int) []uint8 {
	hasher := md5.New()
	hasher.Write([]byte(fmt.Sprintf("%d", num)))
	return hasher.Sum(nil)
}
