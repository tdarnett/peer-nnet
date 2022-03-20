package main

import (
	"context"
	"fmt"
	"log"
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
	echoRPCAPI := EchoRPCAPI{service: s}

	s.rpcServer = rpc.NewServer(s.host, s.protocol)
	err := s.rpcServer.Register(&echoRPCAPI)
	if err != nil {
		return err
	}

	s.rpcClient = rpc.NewClientWithServer(s.host, s.protocol, s.rpcServer)
	return nil
}

func (s *Service) StartMessaging(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// ping available peers every second. Note this uses a different ticket than the dicover peers algo
			s.counter++
			s.Echo(fmt.Sprintf("Message from %s: What's your model version?", s.host.ID().Pretty()))
		}
	}
}

func (s *Service) Echo(message string) {
	peers := FilterSelf(s.host.Peerstore().Peers(), s.host.ID())
	var replies = make([]*Envelope, len(peers))

	errs := s.rpcClient.MultiCall(
		Ctxts(len(peers)),
		peers,
		EchoService,
		EchoServiceFuncEcho,
		Envelope{Message: message},
		CopyEnvelopesToIfaces(replies),
	)

	for i, err := range errs {
		if err != nil {
			fmt.Printf("Peer %s returned error: %-v\n", peers[i].Pretty(), err)
		} else {
			fmt.Printf("Peer %s echoed: %+v\n", peers[i].Pretty(), replies[i].NNet)
		}
	}
}

func (s *Service) ReceiveEcho(envelope Envelope) Envelope {
	// check envelop for incoming model version
	currentModel := new(NeuralNet)
	_, err := s.store.Get(s.hostID, currentModel)

	if err != nil {
		log.Fatal(err)
	}

	return Envelope{
		Message: fmt.Sprintf("Peer %s: %s", s.host.ID(), envelope.Message),
		NNet:    *currentModel,
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

func CopyEnvelopesToIfaces(in []*Envelope) []interface{} {
	ifaces := make([]interface{}, len(in))
	for i := range in {
		in[i] = &Envelope{}
		ifaces[i] = in[i]
	}
	return ifaces
}
