package client

import (
	"errors"
	"sync"
)

// Room 部屋ひとつにつきひとつ Memberたちをまとめる
type Room struct {
	key     string
	members map[int](Member)
	mutex   *sync.Mutex
}

// NewRoom Roomのコンストラクタ
func NewRoom(key string) *Room {
	return &Room{
		key:     key,
		mutex:   &sync.Mutex{},
		members: map[int](Member){},
	}
}

// SetMember メンバーをmapに追加
func (room *Room) SetMember(member Member) error {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	if _, exists := room.members[member.id]; exists {
		return errors.New("already exists")
	}
	room.members[member.id] = member

	return nil
}

// HasMember メンバーを持っているかどうか
func (room *Room) HasMember() bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	return len(room.members) != 0
}

// GetMember メンバーを取得
func (room *Room) GetMember(id int) (Member, bool) {
	room.mutex.Lock()
	m, ok := room.members[id]
	room.mutex.Unlock()
	return m, ok
}

// GetMemberIDList メンバーのIDだけをリスト形式で取得
func (room *Room) GetMemberIDList() []int {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	memberIDList := make([]int, 0, len(room.members))
	for id := range room.members {
		memberIDList = append(memberIDList, id)
	}
	return memberIDList
}

// RemoveMember メンバを削除
func (room *Room) RemoveMember(memberID int) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if _, exists := room.members[memberID]; exists {
		delete(room.members, memberID)
	}
}
