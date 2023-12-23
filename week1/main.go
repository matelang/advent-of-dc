package main

import (
	"crypto/rand"
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

type integerUniqueIDSource struct {
	nodeID  string
	mutex   sync.Mutex
	counter int
}

func NewIntegerUniqueIDSource() UniqueIDSource {
	return &integerUniqueIDSource{
		nodeID:  fmt.Sprintf("%d-%d", os.Getpid(), time.Now().Unix()),
		counter: 0,
	}
}

func (s *integerUniqueIDSource) ID() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.counter++

	return fmt.Sprintf("%s-%d", s.nodeID, s.counter)
}

type randomStringUniqueIDSource struct {
	nodeID string
	chars  string
	length int
}

func NewRandomStringUniqueIDSource(length int) UniqueIDSource {
	var chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

	return &randomStringUniqueIDSource{
		nodeID: fmt.Sprintf("%d-%d", os.Getpid(), time.Now().Unix()),
		chars:  chars,
		length: length,
	}
}

func (s *randomStringUniqueIDSource) ID() string {
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

	n.Handle("generate", generateHandler(n, s))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func generateHandler(n *maelstrom.Node, s UniqueIDSource) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var reply map[string]any

		reply["type"] = "generate_ok"
		reply["id"] = s.ID()
		return n.Reply(msg, reply)
	}
}
