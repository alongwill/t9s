package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/siderolabs/talos/pkg/machinery/resources/block"
	"github.com/siderolabs/talos/pkg/machinery/resources/cluster"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/siderolabs/talos/pkg/machinery/client"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// READ TALOSCONFIG

	talosconfigPath := os.Getenv("TALOSCONFIG")
	if talosconfigPath == "" {
		talosconfigPath = os.ExpandEnv("$HOME/.talos/config")
	}

	log.Printf("using talosconfig: %s", talosconfigPath)

	// BUILD CLIENT

	cli, err := client.New(ctx, client.WithConfigFromFile(talosconfigPath))
	if err != nil {
		return err
	}

	// TARGET MULTIPLE NODES AT THE SAME TIME - MULTIPLEXING

	// ctx = client.WithNodes(ctx, "172.20.0.3", "172.20.0.4", "172.20.0.5")
	ctx = client.WithNodes(ctx, "10.10.8.106")

	// TALOSCTL VERSION:

	version, err := cli.MachineClient.Version(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	for _, message := range version.GetMessages() {
		log.Printf("node: %s version: %s", message.Metadata.Hostname, message.Version.String())
	}

	// TALOSCTL DISKS - THE DEPRECATED WAY

	disksResponse, err := cli.StorageClient.Disks(ctx, &emptypb.Empty{})
	if err != nil {
		return err
	}

	for _, disks := range disksResponse.Messages {
		log.Printf("node: %s disks:", disks.Metadata.Hostname)
		for _, disk := range disks.Disks {
			log.Printf("  disk: %s, size: %d, type: %s", disk.DeviceName, disk.Size, disk.Type.String())
		}
	}

	// COSI EXAMPLES - RESOURCES - TALOSCTL GET ...

	cosiState := cli.COSI

	// COSI state API calls do not support multiplexing (.WithNodes), so we target a single node.
	ctx = client.WithNode(ctx, "10.10.8.106")

	// TALOSCTL GET MEMBERS

	memberList, err := safe.StateListAll[*cluster.Member](ctx, cosiState)
	if err != nil {
		return err
	}

	for member := range memberList.All() {
		log.Printf("got member %q, type: %q", member.Metadata().ID(), member.TypedSpec().MachineType.String())
	}

	// TALOSCTL GET DISKS (THE NEW WAY TO READ DISKS)

	diskList, err := safe.StateListAll[*block.Disk](ctx, cosiState)
	if err != nil {
		return err
	}

	for disk := range diskList.All() {
		log.Printf("got disk %q, transport: %s", disk.Metadata().ID(), disk.TypedSpec().Transport)
	}

	// THE END

	return nil
}
