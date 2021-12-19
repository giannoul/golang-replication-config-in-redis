package redisinfo

import (
	"log"
	"context"
	"fmt"
	"strings"
	"github.com/go-redis/redis/v8"

)

func (r *redisConn) connect() {	
	url := fmt.Sprintf("%s:%d",r.host, r.port)
	r.client = redis.NewClient(&redis.Options{
        Addr:     url,
        Password: "", // no password set
        DB:       0,  // use default DB
    })
}


func CreatePeer(host string, port int, ctx context.Context) (Peer) {
	p := Peer{
		Host: host,
		Port: port,
	}
	p.redisConnInit(ctx)
	return p
}

func (p *Peer) redisConnInit(ctx context.Context) {
	log.Printf("Opening redis connection %s:%d",p.Host,p.Port)
	p.rconn = &redisConn{
		host: p.Host,
		port: p.Port,
		context: ctx,
	}
	p.rconn.connect()
}

func (p *Peer) RedisConnClose() {	
	log.Println("Closing redis connection")
	if err := p.rconn.client.Close(); err != nil{
		log.Fatal(err)
	}	
}


func (p *Peer) Ping() (err error){
	pong, err := p.rconn.client.Ping(p.rconn.client.Context()).Result()
	if err != nil {
		log.Printf("Ping host %s:%d failed with error: %s",p.Host, p.Port, err)
		return err
	}
	log.Printf("Ping host %s:%d successful: %s",p.Host, p.Port, pong)
	return err
}


func (p *Peer) IsMaster() (m bool) {
	m = false 
	info, err := p.rconn.client.Info(p.rconn.client.Context(),"replication").Result()
	if err != nil {
		log.Printf("Info host %s:%d failed with error: %s",p.Host, p.Port, err)
		return false
	}
	temp := strings.Split(info,"\n")[1]
	role := strings.TrimSpace(strings.Split(temp,":")[1])
	log.Printf("Info host %s:%d has role: %s",p.Host, p.Port, role)
	if role == "master" {
		m = true
	}
	return m
}
