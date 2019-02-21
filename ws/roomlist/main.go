package roomlist

import (
	"sync"
	"time"

	"github.com/nerikeshi-k/lemon/ws/client"
)

type roomWithStamp struct {
	room        client.Room
	minLifetime time.Time
}

var (
	rooms = map[string](roomWithStamp){}
	mutex = &sync.Mutex{}
)

const (
	sweepWait = time.Second * 100
)

func init() {
	go SweepEmptyRoom()
}

// GetOrCreateRoom keyに該当するroomを作る なければ作る
func GetOrCreateRoom(key string) client.Room {
	mutex.Lock()
	defer mutex.Unlock()

	r, ok := rooms[key]
	if ok {
		return r.room
	}
	newRoom := *client.NewRoom(key)
	rooms[key] = roomWithStamp{
		room:        newRoom,
		minLifetime: time.Now().Add(time.Second * 10),
	}
	return newRoom
}

// SweepEmptyRoom 空になっている部屋を消す
func SweepEmptyRoom() {
	for {
		time.Sleep(sweepWait)
		mutex.Lock()
		for k, r := range rooms {
			if !r.room.HasMember() && r.minLifetime.Sub(time.Now()) < 0 {
				delete(rooms, k)
			}
		}
		mutex.Unlock()
	}
}
