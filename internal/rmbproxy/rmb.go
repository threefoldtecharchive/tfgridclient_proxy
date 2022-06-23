package rmbproxy

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/threefoldtech/go-rmb"
	"github.com/threefoldtech/go-rmb/client"
)

// MessageIdentifier to get the specific result
type MessageIdentifier struct {
	Retqueue string `json:"retqueue" example:"xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"`
	// TODO: should sign the retqueue and verify it when getting the result to verify the caller
	Signature string `json:"sig,omitempty"`
}

type Rmb struct {
	client *client.MessageBusClient
	ttl    *time.Duration
}

func NewRmb(redis *redis.Client, ttl time.Duration) *Rmb {
	rmbClient := &client.MessageBusClient{
		Client: redis,
		Ctx:    context.TODO(),
	}

	return &Rmb{
		client: rmbClient,
		ttl:    &ttl,
	}
}

func (r *Rmb) envelope(msg *rmb.Message) rmb.Message {
	expiration := msg.Expiration
	if expiration <= 0 {
		expiration = int64(r.ttl.Seconds())
	}

	return rmb.Message{
		Version:    1,
		ID:         "",
		Command:    "system.proxy.request",
		Expiration: expiration,
		Retry:      msg.Retry,
		// data will be overridden by client.Send method for now anyway
		Data:     "",
		TwinSrc:  0,
		TwinDst:  msg.TwinDst,
		Retqueue: uuid.New().String(),
		Schema:   "",
		Epoch:    time.Now().Unix(),
		Err:      "",
	}
}

func (r *Rmb) Submit(msg *rmb.Message) (*MessageIdentifier, error) {
	origMsg, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// TODO: client should hide message construction
	// accepting both a message and a payload (that overrides msg.Data)
	// is a bit confusing
	envelope := r.envelope(msg)
	err = r.client.Send(envelope, string(origMsg))
	if err != nil {
		return nil, err
	}

	return &MessageIdentifier{
		Retqueue: envelope.Retqueue,
	}, nil
}

func (r Rmb) GetResult(msgIdentifier MessageIdentifier) (*[]rmb.Message, error) {
	msg := rmb.Message{
		// only 1 destination for now
		// it's only used by the client to wait
		// for for every destination response
		// so we can use an arbitrary id
		TwinDst:  []int{0},
		Retqueue: msgIdentifier.Retqueue,
	}

	ret := r.client.Read(msg)
	if len(ret) < 1 {
		return nil, errors.New("could not get reply message")
	}

	envelope := ret[0]
	var origMsg rmb.Message

	err := json.Unmarshal([]byte(envelope.Data), &origMsg)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't decode original message")
	}

	return &[]rmb.Message{origMsg}, nil
}
