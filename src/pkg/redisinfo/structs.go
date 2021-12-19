package redisinfo

import (
	"context"
	"github.com/go-redis/redis/v8"
)

type redisConn struct {
	host string
	port int
	context context.Context
	client *redis.Client
}

type Peer struct {
	Host string
	Port int
	rconn *redisConn
}