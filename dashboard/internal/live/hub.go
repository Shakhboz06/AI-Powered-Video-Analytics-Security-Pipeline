package live

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
	"video-analytics-pipe/config"

	"github.com/segmentio/kafka-go"
)

type Hub struct {
	camera map[string]map[chan []byte]struct{}
	mu     sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		camera: make(map[string]map[chan []byte]struct{}),
	}
}

func (h *Hub) Subscribe(cam string) chan []byte {

	
	ch := make(chan []byte, 1)

	
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.camera[cam] == nil {
		h.camera[cam] = make(map[chan []byte]struct{})
	}

	h.camera[cam][ch] = struct{}{}

	return ch

}

func (h *Hub) UnSubcribe(cam string, channel chan []byte) {

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.camera[cam] != nil {
		delete(h.camera[cam], channel)
		close(channel)

		if len(h.camera[cam]) == 0 {
			delete(h.camera, cam)
		}
	}

}

func (h *Hub) Broadcast(cam string, frame []byte) {

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.camera[cam] {
		select {
		case ch <- frame:
		default:
		}

	}
}

func (h *Hub) Run() {

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers:     []string{config.GetString("KAFKA_BROKER", "")},
			Topic:       config.GetString("KAFKA_ANALYSIS_TOPIC", ""),
			GroupID:     fmt.Sprintf("liveview-group-%v", time.Now().UnixMilli()),
			StartOffset: kafka.LastOffset,
		},
	)
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Println("live kafka message error: ", err)
			continue
		}

		latest := map[string][]byte{string(msg.Key): msg.Value}

		for {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
			mes, err := reader.ReadMessage(ctx)
			cancel()
			if err != nil {
				break
			}
			latest[string(mes.Key)] = mes.Value

		}

		for cam, frame := range latest {
			h.Broadcast(cam, frame)
		}
	}

}
