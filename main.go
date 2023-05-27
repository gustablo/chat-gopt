package main

import (
	"fmt"
	"log"

	"github.com/gustablo/chat-gopt/sse"
)

func main() {
	resp, err := sse.GetChatText("give me an example of concurrency code in golang")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Content)
}
