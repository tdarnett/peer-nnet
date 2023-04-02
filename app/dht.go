package main

import (
	"context"
	"log"
	"sync"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
)

// Initializes a new IPFS DHT with the given host and bootstrap peers.
func NewDHT(ctx context.Context, host host.Host, bootstrapPeers []multiaddr.Multiaddr) (*dht.IpfsDHT, error) {
	var options []dht.Option

	// if no bootstrap peers are provided, set the DHT mode to server
	if len(bootstrapPeers) == 0 {
		options = append(options, dht.Mode(dht.ModeServer))
	}

	kdht, err := dht.New(ctx, host, options...)
	if err != nil {
		return nil, err
	}

	if err = kdht.Bootstrap(ctx); err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for _, peerAddr := range bootstrapPeers {
		peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
		if err != nil {
			log.Printf("Failed to parse peer address %s: %s", peerAddr, err)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, *peerInfo); err != nil {
				log.Printf("Error while connecting to node %q: %-v", peerInfo, err)
			} else {
				log.Printf("Connection established with bootstrap node: %q", *peerInfo)
			}
		}()
	}
	wg.Wait()

	return kdht, nil
}
