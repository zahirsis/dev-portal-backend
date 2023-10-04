package messenger

import (
	"github.com/google/uuid"
	"sync"
)

type MessageManager interface {
	Subscribe(ID string) chan []byte
	Unsubscribe(ID string, ch chan []byte)
	Broadcast(ID string, message []byte)
	Close(ID string)
	GenerateID() string
}

type messageManager struct {
	subscriptions map[string][]chan []byte
	mutex         sync.Mutex
}

func NewMessageMassager() MessageManager {
	return &messageManager{
		subscriptions: make(map[string][]chan []byte),
	}
}

func (mm *messageManager) Subscribe(ID string) chan []byte {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	ch := make(chan []byte, 10) // TODO: make buffer size configurable
	mm.subscriptions[ID] = append(mm.subscriptions[ID], ch)

	return ch
}

func (mm *messageManager) Unsubscribe(ID string, ch chan []byte) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	subscriptions, ok := mm.subscriptions[ID]
	if ok {
		delete(mm.subscriptions, ID)
		for i, channel := range subscriptions {
			if channel == ch {
				close(channel)
				mm.subscriptions[ID] = append(subscriptions[:i], subscriptions[i+1:]...)
				break
			}
		}
	}
}

func (mm *messageManager) Broadcast(ID string, message []byte) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	subscribers, ok := mm.subscriptions[ID]
	if ok {
		for _, subscriber := range subscribers {
			go func(ch chan []byte) {
				ch <- message
			}(subscriber)
		}
	}
}

func (mm *messageManager) Close(ID string) {
	for _, subscription := range mm.subscriptions[ID] {
		close(subscription)
	}
	delete(mm.subscriptions, ID)
}

func (mm *messageManager) GenerateID() string {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()
	return uuid.New().String()
}
