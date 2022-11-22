package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/protocol"
	"github.com/multiformats/go-multiaddr"
)

type Config struct {
	Port           int
	ProtocolID     string
	Rendezvous     string
	Seed           int64
	DiscoveryPeers addrList
}

func main() {
	err := godotenv.Load()

	config := Config{}

	flag.StringVar(&config.Rendezvous, "rendezvous", "nnet/echo", "")
	flag.Int64Var(&config.Seed, "seed", 0, "Seed value for generating a PeerID, 0 is random")
	flag.Var(&config.DiscoveryPeers, "peer", "Peer multiaddress for peer discovery")
	flag.StringVar(&config.ProtocolID, "protocolid", "/p2p/rpc/nnet", "")
	flag.IntVar(&config.Port, "port", 0, "")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	initSharedFilesystem()

	h, err := NewHost(ctx, config.Seed, config.Port)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Host ID: %s", h.ID().Pretty())
	log.Printf("Connect to me on:")
	for _, addr := range h.Addrs() {
		log.Printf("  %s/p2p/%s", addr, h.ID().Pretty())
	}

	dht, err := NewDHT(ctx, h, config.DiscoveryPeers)
	if err != nil {
		log.Fatal(err)
	}

	// connect to database
	db, err := InitStore(h.ID().Pretty())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	service := NewService(h, protocol.ID(config.ProtocolID), db)
	err = service.SetupRPC()
	if err != nil {
		log.Fatal(err)
	}

	go Discover(ctx, h, dht, config.Rendezvous)
	go service.StartMessaging(ctx)

	run(h, cancel)
}

func run(h host.Host, cancel func()) {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
	<-c

	fmt.Printf("\rExiting...\n")

	cancel()

	if err := h.Close(); err != nil {
		panic(err)
	}
	os.Exit(0)
}

func initSharedFilesystem() {
	// validate host model information exists
	_, err := os.Stat(HOST_MODEL_WEIGHTS_PATH)
	if err != nil {
		log.Fatal("Failed to locate host model weights file", err)
	}
	_, err = os.Stat(HOST_MODEL_METADATA_PATH)
	if err != nil {
		log.Fatal("Failed to locate host metadata file", err)
	}

	// create peer models directory
	peerModelsDir := filepath.Join(".", PEERS_MODELS_DIR)
	err = MkDir(peerModelsDir)
	if err != nil {
		log.Fatal("Failed to create peers directory", err)
	}

}

type addrList []multiaddr.Multiaddr

func (al *addrList) String() string {
	strs := make([]string, len(*al))
	for i, addr := range *al {
		strs[i] = addr.String()
	}
	return strings.Join(strs, ",")
}

func (al *addrList) Set(value string) error {
	addr, err := multiaddr.NewMultiaddr(value)
	if err != nil {
		return err
	}
	*al = append(*al, addr)
	return nil
}
