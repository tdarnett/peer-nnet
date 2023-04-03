package main

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

// Advertises the given rendezvous string on the network using the provided DHT and attempts 
// to discover and connect to peers that have advertised the same rendezvous string. 
// It runs indefinitely until the context is cancelled or an error occurs.
func Discover(ctx context.Context, h host.Host, dht *dht.IpfsDHT, rendezvous string) {
	var routingDiscovery = discovery.NewRoutingDiscovery(dht)

	discovery.Advertise(ctx, routingDiscovery, rendezvous) // TODO what should TTL be?

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			// try and discover peers every second
			peers, err := discovery.FindPeers(ctx, routingDiscovery, rendezvous)
			if err != nil {
				log.Printf("Failed to find peers: %s", err)
				continue
			}

			// Dial any discovered peers that are not already connected
			for _, peer := range peers {
				if peer.ID == h.ID() {
					continue // filter out self
				}
				if h.Network().Connectedness(peer.ID) != network.Connected {
					_, err = h.Network().DialPeer(ctx, peer.ID)
					if err != nil {
						log.Printf("Failed to dial peer %s: %s", peer.ID.Pretty(), err)
						continue
					}
					log.Printf("Connected to peer %s", peer.ID.Pretty())
				}
			}
		}
	}
}
