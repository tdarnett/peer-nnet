package main

import (
	"context"
	"fmt"
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
	counter   int
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
			// ping available peers every second. Note this uses a different ticket than the dicover peers algo
			s.counter++
			s.RequestVersions(fmt.Sprintf("Message from %s: What's your model version?", s.host.ID().Pretty()))
		}
	}
}

func (s *Service) RequestVersions(message string) {
	peers := FilterSelf(s.host.Peerstore().Peers(), s.host.ID())
	var replies = make([]*ModelVersionContext, len(peers))

	errs := s.rpcClient.MultiCall(
		Ctxts(len(peers)),
		peers,
		NNetService,
		NNetFuncRequestVersion,
		ModelVersionContext{},
		CopyRequestVersionToIfaces(replies),
	)

	var peersWithNewVersions peer.IDSlice
	for i, err := range errs {
		peerId := peers[i].Pretty()
		if err != nil {
			fmt.Printf("Peer %s returned error: %-v\n", peers[i].Pretty(), err)
		} else {
			incomingPeerVersion := replies[i].NNet.version
			fmt.Printf("Peer %s echoed: %+v\n", peerId, replies[i].NNet)
			// compare received version against what the service has internally
			isNewVersion := s.IsNewPeerWeightVersion(peerId, incomingPeerVersion)
			if isNewVersion {
				fmt.Printf("Found new version for Peer: %s. Requesting weights...\n", peerId)
				peersWithNewVersions = append(peersWithNewVersions, peers[i])
				s.UpdatePeerVersion(peerId, incomingPeerVersion)
			}
		}
	}
	// bulk send weight requests
	if len(peersWithNewVersions) > 0 {
		s.RequestModelWeightForPeers(peersWithNewVersions)
	}
}

func (s *Service) ReceiveRequestVersion(requestContext ModelVersionContext) ModelVersionContext {
	// Currently we do not read incoming request contents
	currentModel := new(NeuralNet)
	_, err := s.store.Get(s.hostID, currentModel)

	if err != nil {
		log.Fatal(err)
	}

	return ModelVersionContext{
		Timestamp: time.Now(),
		NNet:      *currentModel,
	}
}

func (s *Service) RequestModelWeightForPeers(peers peer.IDSlice) {
	var replies = make([]*ModelWeightsContext, len(peers))

	errs := s.rpcClient.MultiCall(
		Ctxts(len(peers)),
		peers,
		NNetService,
		NNetFuncRequestModelWeight,
		ModelWeightsContext{},
		CopyRequestWeightsToIfaces(replies),
	)

	for i, err := range errs {
		peerID := peers[i].Pretty()
		if err != nil {
			fmt.Printf("Peer %s returned error: %-v\n", peerID, err)
		} else {
			// create filepath for peer weights
			filename := fmt.Sprintf("%s.h5", peerID)
			peerFilepath := filepath.Join(".", WEIGHTS_DIR, filename)
			WriteFile(peerFilepath, replies[i].Weights)
			fmt.Printf("Peer %s sent their weights. They were saved to: %s\n", filename, peerFilepath)
		}
	}
}

// Returns the current model's weights. Assumes no bad actors.
func (s *Service) ReceiveRequestModelWeight(requestContext ModelWeightsContext) ModelWeightsContext {
	// read weights from file and serialize into context struct
	weightFile := filepath.Join(".", "fixtures", "example-weight.h5") // TODO convert to const
	data, err := os.ReadFile(weightFile)
	if err != nil {
		panic(err)
	}
	return ModelWeightsContext{
		Timestamp: time.Now(),
		Weights:   data,
	}
}

// Determines if an incoming peer nnet is new to receiving service or not
func (s *Service) IsNewPeerWeightVersion(peerID string, incomingVersion int) bool {
	peerModel := new(NeuralNet)
	found, err := s.store.Get(peerID, peerModel)
	if err != nil {
		panic(err)
	}
	isNew := !found

	if found {
		isNew = peerModel.version < incomingVersion
	}

	fmt.Printf("Peer: %s. IsNew: %t\n", peerID, isNew)

	return isNew
}

func (s *Service) UpdatePeerVersion(peerID string, modelVersion int) {
	err := s.store.Set(peerID, NeuralNet{version: modelVersion})
	if err != nil {
		panic(err)
	}
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

func Ctxts(n int) []context.Context {
	ctxs := make([]context.Context, n)
	for i := 0; i < n; i++ {
		ctxs[i] = context.Background()
	}
	return ctxs
}

func CopyRequestVersionToIfaces(in []*ModelVersionContext) []interface{} {
	ifaces := make([]interface{}, len(in))
	for i := range in {
		in[i] = &ModelVersionContext{}
		ifaces[i] = in[i]
	}
	return ifaces
}

func CopyRequestWeightsToIfaces(in []*ModelWeightsContext) []interface{} {
	ifaces := make([]interface{}, len(in))
	for i := range in {
		in[i] = &ModelWeightsContext{}
		ifaces[i] = in[i]
	}
	return ifaces
}
