package main

import (
	"context"
)

const (
	NNetService                = "NNetRPCAPI"
	NNetFuncRequestVersion     = "RequestVersion"
	NNetFuncRequestModelWeight = "RequestModelWeight"
)

type NNetRPCAPI struct {
	service *Service
}

type ModelVersionContext struct {
	Timestamp int64
	Model     NeuralNet
}

type ModelWeightsContext struct {
	Timestamp int64
	Model     NeuralNet
	Weights   []byte // hold h5 filetype. Push to ipfs?
}

func (r *NNetRPCAPI) RequestVersion(ctx context.Context, in ModelVersionContext, out *ModelVersionContext) error {
	// RPC method
	*out = r.service.ReceiveRequestVersion(in)
	return nil
}

func (r *NNetRPCAPI) RequestModelWeight(ctx context.Context, in ModelWeightsContext, out *ModelWeightsContext) error {
	// RPC method
	*out = r.service.ReceiveRequestModelWeight(in)
	return nil
}
