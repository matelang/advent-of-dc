package main

import (
	"log"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("echo", echoHandler(n))

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}

func echoHandler(n *maelstrom.Node) maelstrom.HandlerFunc {
	return func(msg maelstrom.Message) error {
		var reply map[string]any

		reply["type"] = "echo_ok"

		return n.Reply(msg, reply)
	}

}
