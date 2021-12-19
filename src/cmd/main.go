package main

import (
	"fmt"
	"context"
	"strings"
	"strconv"
	"os"
	"log"
	"github.com/giannoul/golang-replication-config-in-redis/pkg/redisinfo"
)

func main() {
	var chans = []chan bool {}
	var res = []bool{}

	inp := os.Getenv("REDISPEERS")
	if inp == "" {
		log.Println("Environment variable REDISPEERS is not set, exiting.")
		os.Exit(1)
	}
	log.Println("REDISPEERS=%s", inp)

	peers := strings.Split(inp,",")
	for i := range peers {
		chans = append(chans, make(chan bool))
		res = append(res, false)
		parts := strings.Split(peers[i],":")
		host := parts[0]
		port,err := strconv.Atoi(parts[1])
		if err != nil {
			log.Println("Error: ",err)
		}
		go worker(host,port,chans[i])
		
	}

	for i := range peers {
		res[i] = <-chans[i]
	}

	master := os.Getenv("HERITAGE")
	masterCount := 0
	for i := range peers {
		if res[i] {
			master = peers[i]
			log.Printf("Master found: %s",master)
			masterCount++
		}
	}

	if masterCount > 1 {
		log.Println("More than 1 masters found, aborting!")
		os.Exit(1)
	}

	if masterCount == 0 && os.Getenv("HERITAGE") == "master" {
		log.Println("I should be the master!")
		os.Exit(0)
	}

	parts := strings.Split(master,":")
	mhost := parts[0]
	mport,err := strconv.Atoi(parts[1])
	if err != nil {
		log.Println("Error: ",err)
	}

	setMasterInRedisConfigFile(mhost, mport)
	log.Printf("%#v",res)
	

	
}

func worker(host string, port int, finished chan bool) {
	log.Println(checkIfHostIsMaster(host, port))
	finished <- checkIfHostIsMaster(host, port)
}

func checkIfHostIsMaster(host string, port int) (r bool) {
	r = false
	p := redisinfo.CreatePeer(host,port,context.Background())

	if p.Ping() != nil {
		return false
	}
	defer p.RedisConnClose()
	return p.IsMaster()
	
}

func setMasterInRedisConfigFile(host string, port int) {
	line := fmt.Sprintf("replicaof %s %d", host, port)
	path := "/etc/redis/redis.conf"
	appendToFile(path, line)
}

func appendToFile(path string, line string) {
	f, err := os.OpenFile(path,os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(line+"\n"); err != nil {
		log.Println(err)
	}
}