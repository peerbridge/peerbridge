package eventbus

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
var Instance = Bus{
	subscriptions: map[string][]Channel{},
	lock:          sync.RWMutex{},
}

func (bus *Bus) Publish(topic string, value interface{}) {
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

func (bus *Bus) Subscribe(topic string) Channel {
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
