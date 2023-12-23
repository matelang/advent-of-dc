package main

import (
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"os"
	"sync"
)

type MessageStore[T comparable] interface {
	Store(message T)
	List() []T
}

type inMemoryMapMessageStore struct {
	mutex    sync.Mutex
	messages map[int]bool
}

func NewInMemoryMessageStore() MessageStore[int] {
	return &inMemoryMapMessageStore{
		messages: map[int]bool{},
	}
}

func (s *inMemoryMapMessageStore) Store(m int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.messages[m] = true
}

func (s *inMemoryMapMessageStore) List() []int {
	keys := make([]int, 0, len(s.messages))
	for k := range s.messages {
		keys = append(keys, k)
	}
	return keys
}

func main() {
	n := maelstrom.NewNode()

	messageStore := NewInMemoryMessageStore()

	n.Handle("broadcast", broadcastHandler(n, messageStore))
	n.Handle("read", readHandler(n, messageStore))
	n.Handle("topology", topologyHandler(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func broadcastHandler(n *maelstrom.Node, s MessageStore[int]) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body struct {
			Message int `json:"message"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		fmt.Fprintf(os.Stderr, "matekaaa received %v", body)

		s.Store(body.Message)

		for _, neighbour := range n.NodeIDs() {
			if neighbour == n.ID() {
				continue
			}

			fmt.Fprintf(os.Stderr, "sending %v to %s", body, neighbour)
			n.Send(neighbour, body)
		}

		reply := map[string]any{}
		reply["type"] = "broadcast_ok"

		return n.Reply(msg, reply)
	}
}

func readHandler(n *maelstrom.Node, s MessageStore[int]) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		messages := s.List()

		body["type"] = "read_ok"
		body["messages"] = messages

		return n.Reply(msg, body)
	}
}

func topologyHandler(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		reply := map[string]any{}
		reply["type"] = "topology_ok"

		return n.Reply(msg, reply)
	}
}
