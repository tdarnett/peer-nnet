package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	rpc "github.com/libp2p/go-libp2p-gorpc"
	"github.com/philippgille/gokv"
)

// Handles the peer-to-peer communication and storage for the application.
type P2PService struct {
	rpcServer *rpc.Server
	rpcClient *rpc.Client
	host      host.Host
	hostID    string
	protocol  protocol.ID
	store     gokv.Store
}

// Creates a new P2PService with the given host, protocol, and store.
func NewP2PService(host host.Host, protocol protocol.ID, store gokv.Store) *P2PService {
	return &P2PService{
		hostID:   host.ID().Pretty(),
		host:     host,
		protocol: protocol,
		store:    store,
	}
}

// Initializes the RPC server and client for the P2PService.
func (s *P2PService) SetupRPC() error {
	nnetRPCAPI := NNetRPCAPI{service: s}

	s.rpcServer = rpc.NewServer(s.host, s.protocol)
	err := s.rpcServer.Register(&nnetRPCAPI)
	if err != nil {
		return err
	}

	s.rpcClient = rpc.NewClientWithServer(s.host, s.protocol, s.rpcServer)
	return nil
}

// Starts the periodic task of requesting model versions from peers.
func (s *P2PService) StartMessaging(ctx context.Context) {
	requestVersionTicker := time.NewTicker(time.Second * 1)
	defer requestVersionTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-requestVersionTicker.C:
			s.RequestVersions()
		}
	}
}

// Sends requests to all known peers to get their current model versions.
func (s *P2PService) RequestVersions() {
	peers := filterSelf(s.host.Peerstore().Peers(), s.host.ID())
	replies := make([]*ModelVersionContext, len(peers))

	errs := s.rpcClient.MultiCall(
		createContexts(len(peers)),
		peers,
		NNetService,
		NNetFuncRequestVersion,
		ModelVersionContext{},
		copyRequestVersionToInterfaces(replies),
	)

	peersWithNewVersions := s.updateStoreWithNewVersions(peers, errs, replies)

	// Bulk requests new weights
	if len(peersWithNewVersions) > 0 {
		s.RequestModelWeights(peersWithNewVersions)
	}
}

// Parses the replies from each peer to identify any new versions, updating its store as necessary
func (s *P2PService) updateStoreWithNewVersions(peers peer.IDSlice, errs []error, replies []*ModelVersionContext) peer.IDSlice {
	var peersWithNewVersions peer.IDSlice

	for i, err := range errs {
		peerID := peers[i].Pretty()
		if err != nil {
			log.Printf("Peer %s returned error: %-v\n", peers[i].Pretty(), err)
		} else {
			incomingPeerModel := replies[i].Model
			log.Printf("Peer %s echoed: %+v\n", peerID, replies[i].Model)

			// Compare received version against the service's internal version
			isNewVersion, err := s.isNewPeerModelVersion(peerID, incomingPeerModel.Version)
			if err != nil {
				log.Print(err) // Log the failure but don't suspend execution
			}
			if isNewVersion {
				log.Printf("Found new version for Peer: %s\n", peerID)
				peersWithNewVersions = append(peersWithNewVersions, peers[i])
				s.updatePeerModel(peerID, incomingPeerModel)
			}
		}
	}

	return peersWithNewVersions
}

// Returns the host's current model version.
// Function triggered from inbound model version request
func (s *P2PService) ReceiveRequestVersion(requestContext ModelVersionContext) ModelVersionContext {
	// Currently we do not read incoming request contents
	currentModel, err := s.getModel()
	if err != nil {
		log.Fatal(err)
	}

	return ModelVersionContext{
		Timestamp: time.Now().Unix(),
		Model:     *currentModel,
	}
}

// Sends a request to the specified peers to obtain their model weights.
func (s *P2PService) RequestModelWeights(peers peer.IDSlice) {
	log.Println("Requesting model weights...")
	var replies = make([]*ModelWeightsContext, len(peers))

	// Multicall will send out the requests AND stream replies
	errs := s.rpcClient.MultiCall(
		createContexts(len(peers)),
		peers,
		NNetService,
		NNetFuncRequestModelWeight,
		ModelWeightsContext{},
		copyRequestWeightsToInterfaces(replies),
	)

	// parse the replies
	for i, err := range errs {
		peerID := peers[i].Pretty()
		if err != nil {
			fmt.Printf("Peer %s returned error: %-v\n", peerID, err)
			continue
		}

		response := replies[i]

		// Create a directory to store the weights received from the peer
		peerDir := filepath.Join(PEERS_MODELS_DIR, peerID)
		if err := os.MkdirAll(peerDir, os.ModePerm); err != nil {
			log.Printf("Error creating directory for peer %s\n", peerID)
			continue
		}

		// Save the weights received from the peer to a file
		weightsFilepath := filepath.Join(peerDir, WEIGHTS_FILENAME)
		if err := WriteFile(weightsFilepath, response.Weights); err != nil {
			log.Printf("unexpected file error: %s\n", err) // fail silently
			continue
		}
		// create filepath for peer model metadata
		metadataFilepath := filepath.Join(peerDir, METADATA_FILENAME)
		content, err := json.Marshal(response.Model)
		if err != nil {
			log.Printf("Error marshaling metadata from peer %s: %s\n", peerID, err)
		}
		if err := WriteFile(metadataFilepath, content); err != nil {
			log.Printf("Unexpected file error saving metadata from peer %s: %s\n", peerID, err)
			continue
		}

		log.Printf("Peer %s sent their weights.\n", peerID)
	}
}

// ReceiveRequestModelWeight returns the host's weights and metadata in response to an inbound model weight request.
// This function assumes that the request is from a trusted peer.
func (s *P2PService) ReceiveRequestModelWeight(requestContext ModelWeightsContext) ModelWeightsContext {
	// Read the weights from file and serialize into the context struct
	data, err := os.ReadFile(HOST_MODEL_WEIGHTS_PATH)
	if err != nil {
		log.Fatalf("Unable to read model weight file %s: %s", HOST_MODEL_WEIGHTS_PATH, err)
	}

	currentModel, err := s.getModel()
	if err != nil {
		log.Fatalf("unable to load model data: %s", err)
	}
	return ModelWeightsContext{
		Timestamp: time.Now().Unix(),
		Model:     *currentModel,
		Weights:   data,
	}
}

// Determines if an incoming peer model is new to the service (hasn't been seen yet)
func (s *P2PService) isNewPeerModelVersion(peerID string, incomingVersion int) (bool, error) {
	peerModel := new(NeuralNet)
	found, err := s.store.Get(peerID, peerModel)
	if err != nil {
		return false, fmt.Errorf("Failed to lookup peer model version from db. Peer %s: %s", peerID, err)
	}
	isNew := !found

	if found {
		isNew = peerModel.Version < incomingVersion
	}

	return isNew, nil
}

// Updates a provided peer's model information
func (s *P2PService) updatePeerModel(peerID string, model NeuralNet) error {
	err := s.store.Set(peerID, model)
	if err != nil {
		return fmt.Errorf("unable to set peer %s model version to %d: %s", peerID, model.Version, err)
	}
	return nil
}

// Fetches and returns a hosts own model metadata from disk
func (s *P2PService) getModel() (*NeuralNet, error) {
	currentModel := NeuralNet{}
	file, err := ioutil.ReadFile(HOST_MODEL_METADATA_PATH)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(file), &currentModel)

	return &currentModel, nil
}

func filterSelf(peers peer.IDSlice, self peer.ID) peer.IDSlice {
	var withoutSelf peer.IDSlice
	for _, p := range peers {
		if p != self {
			withoutSelf = append(withoutSelf, p)
		}
	}
	return withoutSelf
}

func createContexts(n int) []context.Context {
	ctxs := make([]context.Context, n)
	for i := 0; i < n; i++ {
		ctxs[i] = context.Background()
	}
	return ctxs
}

func copyRequestVersionToInterfaces(in []*ModelVersionContext) []interface{} {
	ifaces := make([]interface{}, len(in))
	for i := range in {
		in[i] = &ModelVersionContext{}
		ifaces[i] = in[i]
	}
	return ifaces
}

func copyRequestWeightsToInterfaces(in []*ModelWeightsContext) []interface{} {
	ifaces := make([]interface{}, len(in))
	for i := range in {
		in[i] = &ModelWeightsContext{}
		ifaces[i] = in[i]
	}
	return ifaces
}
