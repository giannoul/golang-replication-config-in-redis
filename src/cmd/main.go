package main

import (
	"fmt"
	"context"
	"strings"
	"strconv"
	"os"
	"log"
	"io/ioutil"
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
		host, port := ipPortPairToParts(peers[i])
		go worker(host,port,chans[i])
		
	}

	for i := range peers {
		res[i] = <-chans[i]
	}

	master := getEnv("EXPOSING_SVC","")
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

	mhost, mport := "", 0

	switch masterCount {
    case 0:
        if os.Getenv("HERITAGE") == "master" {
			log.Println("I should be the master! My details are: ", master)
			mhost, mport = ipPortPairToParts(master)
			copyFile(getEnv("REDIS_CONF_SRC_PATH", "/etc/redis/vanilla-redis.conf"), getEnv("REDIS_CONF_DST_PATH", "/etc/redis/redis.conf")) 
		}
    case 1:
        mhost, mport = ipPortPairToParts(master)
		setMasterInRedisConfigFile(mhost, mport)
    default:
        log.Fatal("I shouldn't even get here!")
		os.Exit(1)
    }
	
	setMasterInSentinelConfigFile(mhost, mport)
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
	sourcePath := getEnv("REDIS_CONF_SRC_PATH", "/etc/redis/vanilla-redis.conf")
	destinationPath := getEnv("REDIS_CONF_DST_PATH", "/etc/redis/redis.conf")
	copyFile(sourcePath, destinationPath)
	appendToFile(destinationPath, line)
}

func setMasterInSentinelConfigFile(host string, port int) {
	line := fmt.Sprintf("sentinel monitor mymaster %s %d 2", host, port)
	sourcePath := getEnv("SENTINEL_CONF_SRC_PATH", "/etc/redis/vanilla-sentinel.conf")
	destinationPath := getEnv("SENTINEL_CONF_DST_PATH", "/etc/redis/sentinel.conf")
	copyFile(sourcePath, destinationPath)
	appendToFile(destinationPath, line)
}

func appendToFile(path string, line string) {
	f, err := os.OpenFile(path,os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	log.Println(line)
	if _, err := f.WriteString("\n"+line+"\n"); err != nil {
		log.Println(err)
	}
}

func copyFile(src string, dest string) {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		log.Println(err)
			return
	}

	err = ioutil.WriteFile(dest, input, 0666)
	if err != nil {
		log.Println("Error creating", dest)
		log.Fatal(err)
		return
	}
}

func getEnv(key, fallback string) string {
    if value, ok := os.LookupEnv(key); ok {
        return value
    }
    return fallback
}

func ipPortPairToParts(pair string) (host string, port int) {
	parts := strings.Split(pair,":")
	host = parts[0]
	port,err := strconv.Atoi(parts[1])
	if err != nil {
		log.Fatal("Error: ",err)
	}
	return host, port
}