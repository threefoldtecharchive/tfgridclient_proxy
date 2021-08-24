package explorer

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// SetRedisKey to Get conn and put back when exit from method
func (a *App) SetRedisKey(key string, val []byte, expiration uint64) error {
	conn := a.redis.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, val, "EX", expiration)
	if err != nil {
		log.Error().Err(errors.Wrap(err, fmt.Sprintf("ERROR: fail set key %s, val %s", key, val))).Msg("")
		return err
	}

	return nil
}

// GetRedisKey to Get conn and put back when exit from method
func (a *App) GetRedisKey(key string) (string, error) {
	conn := a.redis.Get()
	defer conn.Close()

	s, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.Error().Err(errors.Wrap(err, fmt.Sprintf("ERROR: fail get key %s", key))).Msg("")
		return "", err
	}

	return s, nil
}
