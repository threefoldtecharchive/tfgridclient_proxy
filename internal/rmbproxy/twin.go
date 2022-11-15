package rmbproxy

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"

	"github.com/threefoldtech/go-rmb"
	"github.com/threefoldtech/substrate-client"
)

// TwinExplorerResolver is Substrate resolver
type TwinResolver struct {
	client *substrate.Substrate
	cache  *cache.Cache
	ttl    time.Duration
}

// NewTwinResolver : create a new substrate resolver
func NewTwinResolver(substrateURL string, redis *redis.Client, ttl time.Duration) (*TwinResolver, error) {
	client, err := substrate.NewSubstrate(substrateURL)
	if err != nil {
		return nil, err
	}

	redisCache := cache.New(&cache.Options{
		Redis:      redis,
		LocalCache: cache.NewTinyLFU(1000, ttl),
	})

	return &TwinResolver{
		client: client,
		cache:  redisCache,
		ttl:    ttl,
	}, nil
}

func (r TwinResolver) Get(id int) (*substrate.Twin, error) {
	var twin *substrate.Twin

	ctx := context.TODO()
	key := strconv.Itoa(id)
	if err := r.cache.Get(ctx, key, &twin); err == nil {
		return twin, nil
	}

	twin, err := r.client.GetTwin(uint32(id))
	if err != nil {
		return nil, err
	}

	if err := r.cache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: twin,
		TTL:   r.ttl,
	}); err != nil {
		return nil, err
	}

	return twin, nil
}

func (r TwinResolver) Verify(id int, message *rmb.Message) error {
	twin, err := r.Get(id)
	if err != nil {
		return err
	}

	pubKey := twin.Account.PublicKey()
	return message.Verify(pubKey)
}
