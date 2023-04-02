package main

import (
	"path/filepath"

	"github.com/libp2p/go-libp2p-core/peer"
)

// PeerConfig contains the configuration for a peer, including its ID and filepaths for its model weights and metadata.
type PeerConfig struct {
	ID          peer.ID
	PeerDir			string
	MetadataDir string
	WeightsDir  string
}

// Creates a new PeerConfig for a given peer ID.
func NewPeerConfig(id peer.ID) *PeerConfig {
	peerDir := filepath.Join(peerModelDir, id.Pretty())
	return &PeerConfig{
		ID:          id,
		PeerDir:     peerDir,
		MetadataDir: filepath.Join(peerDir, METADATA_FILENAME),
		WeightsDir:  filepath.Join(peerDir, WEIGHTS_FILENAME),
	}
}
