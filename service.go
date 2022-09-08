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

type Service struct {
	rpcServer *rpc.Server
	rpcClient *rpc.Client
	host      host.Host
	hostID    string
	protocol  protocol.ID
	store     gokv.Store
}

func NewService(host host.Host, protocol protocol.ID, store gokv.Store) *Service {
	return &Service{
		hostID:   host.ID().Pretty(), // helpful to access easily
		host:     host,
		protocol: protocol,
		store:    store,
	}
}

func (s *Service) SetupRPC() error {
	nnetRPCAPI := NNetRPCAPI{service: s}

	s.rpcServer = rpc.NewServer(s.host, s.protocol)
	err := s.rpcServer.Register(&nnetRPCAPI)
	if err != nil {
		return err
	}

	s.rpcClient = rpc.NewClientWithServer(s.host, s.protocol, s.rpcServer)
	return nil
}

func (s *Service) StartMessaging(ctx context.Context) {
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

func (s *Service) RequestVersions() {
	peers := FilterSelf(s.host.Peerstore().Peers(), s.host.ID())
	var replies = make([]*ModelVersionContext, len(peers))

	errs := s.rpcClient.MultiCall(
		ctxts(len(peers)),
		peers,
		NNetService,
		NNetFuncRequestVersion,
		ModelVersionContext{},
		copyRequestVersionToIfaces(replies),
	)

	var peersWithNewVersions peer.IDSlice

	for i, err := range errs {
		peerID := peers[i].Pretty()
		if err != nil {
			fmt.Printf("Peer %s returned error: %-v\n", peers[i].Pretty(), err)
		} else {
			incomingPeerNNet := replies[i].Model
			fmt.Printf("Peer %s echoed: %+v\n", peerID, replies[i].Model)

			// compare received version against what the service has internally
			isNewVersion, err := s.isNewPeerModelVersion(peerID, incomingPeerNNet.Version)
			if err != nil {
				fmt.Print(err) // log the failure but don't suspend execution
			}
			if isNewVersion {
				fmt.Printf("Found new version for Peer: %s\n", peerID)
				peersWithNewVersions = append(peersWithNewVersions, peers[i])
				s.updatePeerModel(peerID, incomingPeerNNet)
			}
		}
	}
	// bulk send weight requests
	if len(peersWithNewVersions) > 0 {
		s.RequestModelWeights(peersWithNewVersions)
	}
}

// Returns the host's current model version.
// Function triggered from inbound model version request
func (s *Service) ReceiveRequestVersion(requestContext ModelVersionContext) ModelVersionContext {
	// Currently we do not read incoming request contents
	// TODO read model metadata from model/ path that the ml model training process is writing to
	currentModel, err := s.getModel()
	if err != nil {
		log.Fatal(err)
	}

	return ModelVersionContext{
		Timestamp: time.Now().Unix(),
		Model:     *currentModel,
	}
}

func (s *Service) RequestModelWeights(peers peer.IDSlice) {
	fmt.Printf("Requesting model weights...")
	var replies = make([]*ModelWeightsContext, len(peers))

	// Multicall will send out the requests AND stream replies
	errs := s.rpcClient.MultiCall(
		ctxts(len(peers)),
		peers,
		NNetService,
		NNetFuncRequestModelWeight,
		ModelWeightsContext{},
		copyRequestWeightsToIfaces(replies),
	)

	// parse the replies
	for i, err := range errs {
		peerID := peers[i].Pretty()
		if err != nil {
			fmt.Printf("Peer %s returned error: %-v\n", peerID, err)
			continue
		}
		peerResponse := replies[i]
		// create peer directory
		peerDir := filepath.Join(".", PEER_MODELS_DIR, peerID)
		MkDir(peerDir)
		if err != nil {
			fmt.Printf("error creating peer directory\n")
			continue
		}

		// create filepath for peer weights
		weightsFilepath := filepath.Join(peerDir, "weights.h5")
		err = WriteFile(weightsFilepath, peerResponse.Weights)
		if err != nil {
			fmt.Printf("unexpected file error: %s\n", err) // fail silently
			continue
		}
		// create filepath for peer model metadata
		metadataFilepath := filepath.Join(peerDir, "metadata.json")
		content, err := json.Marshal(peerResponse.Model)
		if err != nil {
			fmt.Println(err)
		}
		err = WriteFile(metadataFilepath, content)
		if err != nil {
			fmt.Printf("unexpected file error: %s\n", err) // fail silently
			continue
		}

		fmt.Printf("Peer %s sent their weights. They were saved to: %s\n", peerID, peerDir)
	}
}

// Returns a host's weights and metadata.
// Function triggered from inbound model weight request.
// Assumes no bad actors.
func (s *Service) ReceiveRequestModelWeight(requestContext ModelWeightsContext) ModelWeightsContext {
	// read weights from file and serialize into context struct
	weightFile := filepath.Join(".", "model", "weights.h5")
	data, err := os.ReadFile(weightFile)
	if err != nil {
		log.Fatalf("unable to read model weight file %s: %s", weightFile, err)
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
func (s *Service) isNewPeerModelVersion(peerID string, incomingVersion int) (bool, error) {
	peerModel := new(NeuralNet)
	found, err := s.store.Get(peerID, peerModel)
	if err != nil {
		return false, fmt.Errorf("failed to lookup peer model version from db. Peer %s: %s", peerID, err)
	}
	isNew := !found

	if found {
		isNew = peerModel.Version < incomingVersion
	}

	fmt.Printf("Peer: %s. IsNew: %t\n", peerID, isNew) // TODO remove

	return isNew, nil
}

// Updates a provided peer's model information
func (s *Service) updatePeerModel(peerID string, model NeuralNet) error {
	err := s.store.Set(peerID, model)
	if err != nil {
		return fmt.Errorf("unable to set peer %s model version to %d: %s", peerID, model.Version, err)
	}
	return nil
}

// Fetches and returns a hosts own model metadata from disk
func (s *Service) getModel() (*NeuralNet, error) {
	currentModel := NeuralNet{}
	metadataFile := filepath.Join(".", "model", "metadata.json")
	file, err := ioutil.ReadFile(metadataFile)

	if err != nil {
		return nil, err
	}

	json.Unmarshal([]byte(file), &currentModel)

	return &currentModel, nil
}

func FilterSelf(peers peer.IDSlice, self peer.ID) peer.IDSlice {
	var withoutSelf peer.IDSlice
	for _, p := range peers {
		if p != self {
			withoutSelf = append(withoutSelf, p)
		}
	}
	return withoutSelf
}

func ctxts(n int) []context.Context {
	ctxs := make([]context.Context, n)
	for i := 0; i < n; i++ {
		ctxs[i] = context.Background()
	}
	return ctxs
}

func copyRequestVersionToIfaces(in []*ModelVersionContext) []interface{} {
	ifaces := make([]interface{}, len(in))
	for i := range in {
		in[i] = &ModelVersionContext{}
		ifaces[i] = in[i]
	}
	return ifaces
}

func copyRequestWeightsToIfaces(in []*ModelWeightsContext) []interface{} {
	ifaces := make([]interface{}, len(in))
	for i := range in {
		in[i] = &ModelWeightsContext{}
		ifaces[i] = in[i]
	}
	return ifaces
}
