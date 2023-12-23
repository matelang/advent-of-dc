package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"os"
	"sync"
	"time"
)

type UniqueIDSource interface {
	ID() string
}

type IntegerUniqueIDSource struct {
	nodeID  string
	mutex   sync.Mutex
	counter int
}

func NewIntegerUniqueIDSource() UniqueIDSource {
	return &IntegerUniqueIDSource{
		nodeID:  fmt.Sprintf("%d-%d", os.Getpid(), time.Now().Unix()),
		counter: 0,
	}
}

func (s *IntegerUniqueIDSource) ID() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.counter++

	return fmt.Sprintf("%s-%d", s.nodeID, s.counter)
}

type RandomStringUniqueIDSource struct {
	nodeID string
	chars  string
	length int
}

func NewRandomStringUniqueIDSource(length int) UniqueIDSource {
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	return &RandomStringUniqueIDSource{
		nodeID: fmt.Sprintf("%d-%d", os.Getpid(), time.Now().Unix()),
		chars:  chars,
		length: length,
	}
}

func (s *RandomStringUniqueIDSource) ID() string {
	ll := len(s.chars)
	b := make([]byte, s.length)
	rand.Read(b)
	for i := 0; i < s.length; i++ {
		b[i] = s.chars[int(b[i])%ll]
	}

	return fmt.Sprintf("%s-%s", s.nodeID, string(b))
}

func main() {
	n := maelstrom.NewNode()

	//s := NewIntegerUniqueIDSource()
	s := NewRandomStringUniqueIDSource(13)

	n.Handle("generate", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		body["type"] = "generate_ok"
		body["id"] = s.ID()
		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
