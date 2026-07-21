package live

import (
	"context"
	"log"
	"sync"
	"video-analytics-pipe/config"

	"github.com/segmentio/kafka-go"
)

type Hub struct {
	camera map[string]map[chan[]byte]struct{}
	mu     sync.RWMutex
}


func NewHub() *Hub{
	return &Hub{
		camera: make(map[string]map[chan []byte]struct{}),
	}
}

func (h *Hub) Subscribe (cam string) chan[]byte{

	//buffered, not unbuffered. make(chan []byte, 2) not make(chan []byte). 
	//The buffer is what makes "drop if slow" possible later: Broadcast's non-blocking send fills the buffer, 
	// and only drops when it's full. An unbuffered channel would drop every frame unless a reader is waiting at the exact instant — useless here. 
	// The buffer size is the knob: 2–4 frames.
	//buffer_size is set to 4 here.
	ch := make(chan []byte, 4)


	// writing to the map (adding a channel, maybe creating a set). 
	// That's a mutation, so it needs the exclusive write lock, even though it feels quick. 
	// RLock is only for Broadcast, which reads. defer unlock() right after locking is the clean 
	// Go habit — you can't forget to release it.
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.camera[cam] == nil{
		h.camera[cam] = make(map[chan []byte]struct{})
	}

	h.camera[cam][ch] = struct{}{}

	return ch

}



func (h *Hub) UnSubcribe(cam string, channel chan[]byte){
	
	// just deletetion operation overall this operation
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.camera[cam] != nil{
		delete(h.camera[cam], channel)
		close(channel)

		if len(h.camera[cam]) == 0{
			delete(h.camera, cam)
		}	
	} 

}

func(h *Hub) Broadcast(cam string, frame []byte){

	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.camera[cam]{
		select{
			case ch <- frame:
			default:
		}

	}
}


func(h * Hub) Run (){

	reader := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers: []string{config.GetString("KAFKA_BROKER", "")},
			Topic: config.GetString("KAFKA_ANALYSIS_TOPIC", ""),
			GroupID: "liveview-group",
			StartOffset: kafka.LastOffset,
		},		
	)
	defer reader.Close()

	for{
		msg, err := reader.ReadMessage(context.Background())
		if err != nil{
			log.Println("live kafka message error: ", err)
			continue
		}

		camera := string(msg.Key)

		frame := msg.Value

		h.Broadcast(camera, frame)

		// log.Printf("broadcast %s (%d bytes)", camera, len(frame))
	}

}