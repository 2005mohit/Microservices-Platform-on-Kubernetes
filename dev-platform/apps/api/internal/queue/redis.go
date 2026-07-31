package queue

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client
var ctx = context.Background()
var subs = struct {
	sync.RWMutex
	m map[string][]chan string
}{m: make(map[string][]chan string)}

func Init(addr string) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return err
	}
	go listenPubSub()
	return nil
}

func Close() {
	if rdb != nil {
		rdb.Close()
	}
}

func Publish(channel string, data interface{}) {
	msg, err := json.Marshal(data)
	if err != nil {
		log.Printf("redis marshal error: %v", err)
		return
	}
	if rdb != nil {
		rdb.Publish(ctx, channel, string(msg))
	}
	subs.RLock()
	for _, ch := range subs.m[channel] {
		select {
		case ch <- string(msg):
		default:
		}
	}
	subs.RUnlock()
}

func Subscribe(channel string, ch chan string) int {
	subs.Lock()
	subs.m[channel] = append(subs.m[channel], ch)
	id := len(subs.m[channel]) - 1
	subs.Unlock()
	return id
}

func Unsubscribe(channel string, id int) {
	subs.Lock()
	if subs.m[channel] != nil && id < len(subs.m[channel]) {
		close(subs.m[channel][id])
		subs.m[channel] = append(subs.m[channel][:id], subs.m[channel][id+1:]...)
	}
	subs.Unlock()
}

func listenPubSub() {
	if rdb == nil {
		return
	}
	pubsub := rdb.PSubscribe(ctx, "deployment:*")
	defer pubsub.Close()

	ch := pubsub.Channel()
	for msg := range ch {
		subs.RLock()
		for _, ch := range subs.m[msg.Channel] {
			select {
			case ch <- msg.Payload:
			default:
			}
		}
		subs.RUnlock()
	}
}
