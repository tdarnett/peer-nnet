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

type AppConfig struct {
	Port           int
	ProtocolID     string
	Rendezvous     string
	Seed           int64
	DiscoveryPeers addrList
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	config := AppConfig{}

	// Define command line flags
	flag.StringVar(&config.Rendezvous, "rendezvous", "nnet/echo", "Rendezvous string for peer discovery")
	flag.Int64Var(&config.Seed, "seed", 0, "Seed value for generating a PeerID, 0 is random")
	flag.Var(&config.DiscoveryPeers, "peer", "Peer multiaddress for peer discovery")
	flag.StringVar(&config.ProtocolID, "protocolid", "/p2p/rpc/nnet", "Protocol ID for communication between peers")
	flag.IntVar(&config.Port, "port", 0, "Port number to listen on")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	// Initialize shared filesystem
	initSharedFilesystem()

	h, err := NewHost(ctx, config.Seed, config.Port)
	if err != nil {
		log.Fatalf("Error creating host: %v", err)
	}

	log.Printf("Host ID: %s", h.ID().Pretty())
	log.Printf("Connect to me on:")
	for _, addr := range h.Addrs() {
		log.Printf("  %s/p2p/%s", addr, h.ID().Pretty())
	}

	dht, err := NewDHT(ctx, h, config.DiscoveryPeers)
	if err != nil {
		log.Fatalf("Error creating DHT: %v", err)
	}

	// connect to database
	db, err := InitStore(h.ID().Pretty())
	if err != nil {
		log.Fatalf("Error initializing store: %v", err)
	}
	defer db.Close()

	service := NewService(h, protocol.ID(config.ProtocolID), db)
	err = service.SetupRPC()
	if err != nil {
		log.Fatalf("Error setting up RPC: %v", err)
	}

	// Start discovery and messaging services
	go Discover(ctx, h, dht, config.Rendezvous)
	go service.StartMessaging(ctx)

	// Wait for termination signal and perform cleanup
	waitForTermination(h, cancel)
}

// waitForTermination waits for a termination signal, then cancels the context, closes the host, and exits the program.
func waitForTermination(h host.Host, cancel func()) {
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
	// Validate host model information exists
	if err := checkFileExists(HOST_MODEL_WEIGHTS_PATH); err != nil {
		log.Fatalf("Failed to locate host model weights file: %v", err)
	}

	if err := checkFileExists(HOST_MODEL_METADATA_PATH); err != nil {
		log.Fatalf("Failed to locate host metadata file: %v", err)
	}

	// Create peer models directory
	peerModelsDir := filepath.Join(".", PEERS_MODELS_DIR)
	if err := os.MkdirAll(peerModelsDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create peers directory: %v", err)
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
		return fmt.Errorf("error parsing multiaddress: %v", err)
	}
	*al = append(*al, addr)
	return nil
}
