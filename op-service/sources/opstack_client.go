package sources

import (
	"context"

	"github.com/ethereum-optimism/optimism/op-service/client"
	"github.com/ethereum-optimism/optimism/op-service/eth"
)

type BuildAPI interface {
	// OpenBlock starts a block-building job with the given attributes.
	// The identifier of the job is returned, if successfully started.
	OpenBlock(ctx context.Context, parent eth.BlockID, attrs *eth.PayloadAttributes) (eth.PayloadInfo, error)
	// CancelBlock cancels block-building.
	CancelBlock(ctx context.Context, id eth.PayloadInfo) error
	// SealBlock completes block-building. The block will not be canonical until committed to by CommitBlock
	SealBlock(ctx context.Context, id eth.PayloadInfo) (*eth.ExecutionPayloadEnvelope, error)
}

type CommitAPI interface {
	// CommitBlock processes the block, and sets it as canonical block of the chain.
	CommitBlock(ctx context.Context, envelope *eth.SignedExecutionPayloadEnvelope) error
}

type PublishAPI interface {
	PublishBlock(ctx context.Context, signed *eth.SignedExecutionPayloadEnvelope) error
}

type OPStackAPI interface {
	BuildAPI
	CommitAPI
	PublishAPI
}

type OPStackClient struct {
	rpc client.RPC
}

func NewOPStackClient(rpc client.RPC) *OPStackClient {
	return &OPStackClient{rpc}
}

func (r *OPStackClient) OpenBlock(ctx context.Context, parent eth.BlockID, attrs *eth.PayloadAttributes) (eth.PayloadInfo, error) {
	var result eth.PayloadInfo
	err := r.rpc.CallContext(ctx, &result, "opstack_openBlockV1", parent, attrs)
	return result, err
}

func (r *OPStackClient) CancelBlock(ctx context.Context, id eth.PayloadInfo) error {
	return r.rpc.CallContext(ctx, nil, "opstack_cancelBlockV1", id)
}

func (r *OPStackClient) SealBlock(ctx context.Context, id eth.PayloadInfo) (*eth.ExecutionPayloadEnvelope, error) {
	var result *eth.ExecutionPayloadEnvelope
	err := r.rpc.CallContext(ctx, &result, "opstack_sealBlockV1", id)
	return result, err
}

func (r *OPStackClient) CommitBlock(ctx context.Context, envelope *eth.SignedExecutionPayloadEnvelope) error {
	return r.rpc.CallContext(ctx, nil, "opstack_commitBlockV1", envelope)
}

func (r *OPStackClient) PublishBlock(ctx context.Context, signed *eth.SignedExecutionPayloadEnvelope) error {
	return r.rpc.CallContext(ctx, nil, "opstack_publishBlockV1", signed)
}

func (r *OPStackClient) Close() {
	r.rpc.Close()
}
