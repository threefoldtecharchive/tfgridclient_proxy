package explorer

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// SetRedisKey to Get conn and put back when exit from method
func (a *App) SetRedisKey(key string, val []byte, expiration uint64) error {
	conn := a.redis.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, val, "EX", expiration)
	if err != nil {
		return fmt.Errorf("failed set key %s, val %s : %w", key, val, err)
	}

	return nil
}

// GetRedisKey to Get conn and put back when exit from method
func (a *App) GetRedisKey(key string) (string, error) {
	conn := a.redis.Get()
	defer conn.Close()

	s, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", fmt.Errorf("failed get key %s : %w", key, err)
	}

	return s, nil
}
