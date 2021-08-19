package explorer

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

var pool *redis.Pool

func InitRedisPool() {
	pool = &redis.Pool{
		MaxIdle:   10,
		MaxActive: 10,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", "localhost:6379")
			if err != nil {
				log.Error().Err(errors.Wrap(err, "ERROR: fail init redis")).Msg("")
			}
			return conn, err
		},
	}
}

func SetRedisKey(key string, val []byte, expiration uint64) error {
	// get conn and put back when exit from method
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, val, "EX", expiration)
	if err != nil {
		log.Error().Err(errors.Wrap(err, fmt.Sprintf("ERROR: fail set key %s, val %s", key, val))).Msg("")
		return err
	}

	return nil
}

func GetRedisKey(key string) (string, error) {
	// get conn and put back when exit from method
	conn := pool.Get()
	defer conn.Close()

	s, err := redis.String(conn.Do("GET", key))
	if err != nil {
		log.Error().Err(errors.Wrap(err, fmt.Sprintf("ERROR: fail get key %s", key))).Msg("")
		return "", err
	}

	return s, nil
}
