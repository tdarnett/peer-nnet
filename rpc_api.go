package main

import (
	"context"
)

const (
	EchoService         = "EchoRPCAPI"
	EchoServiceFuncEcho = "Echo"
)

type EchoRPCAPI struct {
	service *Service
}

type Envelope struct {
	Message string
	NNet    NeuralNet
}

func (r *EchoRPCAPI) Echo(ctx context.Context, in Envelope, out *Envelope) error {
	// RPC method
	*out = r.service.ReceiveEcho(in)
	return nil
}
