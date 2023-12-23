package main

import (
	"encoding/json"
	"fmt"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
	"log"
	"os"
	"slices"
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

var neighborListMutex sync.Mutex
var neighborList []string

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
			Type    string `json:"type"`
			Message int    `json:"message"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		s.Store(body.Message)

		for _, neighbour := range neighborList {
			if neighbour == msg.Src {
				continue
			}

			fmt.Fprintf(os.Stderr, "mate - sending %v to %s\n", body, neighbour)
			go func(nn string) {
				err := n.RPC(nn, body, func(msg maelstrom.Message) error {
					if msg.Type() != "broadcast_ok" {
						log.Fatal("unexpected brodacast message response type %s", msg.Type())
					}

					return nil
				})

				if err != nil {
					log.Fatal(
						fmt.Errorf("error sending RPC while trying to broadcast message %v to neighbors: %w",
							msg.Body, err),
					)
				}
			}(neighbour)
		}

		reply := map[string]any{}
		reply["type"] = "broadcast_ok"

		return n.Reply(msg, reply)
	}
}

func readHandler(n *maelstrom.Node, s MessageStore[int]) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var reply struct {
			Type     string `json:"type"`
			Messages []int  `json:"messages"`
		}

		reply.Type = "read_ok"
		reply.Messages = s.List()

		fmt.Fprintf(os.Stderr, "mate - returning to read from %s message %v\n", msg.Src, reply.Messages)

		return n.Reply(msg, reply)
	}
}

func topologyHandler(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var body struct {
			Topology map[string][]string `json:"topology"`
		}
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		neighborListMutex.Lock()
		defer neighborListMutex.Unlock()

		updatedNeighbors := body.Topology[n.ID()]

		slices.Sort(updatedNeighbors)

		var newNeighbors []string

		neighborList = updatedNeighbors

		reply := map[string]any{}
		reply["type"] = "topology_ok"

		return n.Reply(msg, reply)
	}
}
