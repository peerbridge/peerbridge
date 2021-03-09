package blockchain

import "sync"

type Event struct {
	Topic string
	Value interface{}
}

type Channel chan Event

type Bus struct {
	subscriptions map[string][]Channel
	lock          sync.RWMutex
}

// The main event bus instance.
var EventBus = Bus{
	subscriptions: map[string][]Channel{},
	lock:          sync.RWMutex{},
}

const (
	newLocalBlockTopic        = "peerbridge.topics.local.new-block"
	newRemoteBlockTopic       = "peerbridge.topics.remote.new-block"
	newLocalTransactionTopic  = "peerbridge.topics.local.new-transaction"
	newRemoteTransactionTopic = "peerbridge.topics.remote.new-transaction"
)

func (bus *Bus) PublishNewLocalBlock(b Block) {
	bus.publish(newLocalBlockTopic, b)
}

func (bus *Bus) PublishNewRemoteBlock(b Block) {
	bus.publish(newRemoteBlockTopic, b)
}

func (bus *Bus) PublishNewLocalTransaction(t Transaction) {
	bus.publish(newLocalTransactionTopic, t)
}

func (bus *Bus) PublishNewRemoteTransaction(t Transaction) {
	bus.publish(newRemoteTransactionTopic, t)
}

func (bus *Bus) SubscribeNewLocalBlock() Channel {
	return bus.subscribe(newLocalBlockTopic)
}

func (bus *Bus) SubscribeNewRemoteBlock() Channel {
	return bus.subscribe(newRemoteBlockTopic)
}

func (bus *Bus) SubscribeNewLocalTransaction() Channel {
	return bus.subscribe(newLocalTransactionTopic)
}

func (bus *Bus) SubscribeNewRemoteTransaction() Channel {
	return bus.subscribe(newRemoteTransactionTopic)
}

func (bus *Bus) publish(topic string, value interface{}) {
	bus.lock.RLock()
	if chans, found := bus.subscriptions[topic]; found {
		// Create a slice snapshot to unlock it after
		// the concurrent exection of the publishing
		channels := append([]Channel{}, chans...)
		publishAsync := func(event Event, channels []Channel) {
			for _, channel := range channels {
				channel <- event
			}
		}
		go publishAsync(Event{Topic: topic, Value: value}, channels)
	}
	bus.lock.RUnlock()
}

func (bus *Bus) subscribe(topic string) Channel {
	channel := make(chan Event)
	bus.lock.Lock()
	if prev, found := bus.subscriptions[topic]; found {
		bus.subscriptions[topic] = append(prev, channel)
	} else {
		bus.subscriptions[topic] = append([]Channel{}, channel)
	}
	bus.lock.Unlock()
	return channel
}
