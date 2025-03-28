package main

import (
	"context"
	"log"

	"github.com/siderolabs/talos/pkg/machinery/client"
)

func main() {
	println("hello")

	log.Printf("sdfdsfdsf")

	ctx := client.WithNode(context.Background(), "172.20.0.3")

	client, err := client.New(ctx)
	if err != nil {
		panic(err)
	}

	_ = client
}
