package main

import (
	"github.com/gomodule/redigo/redis"
	"log"
)

var pool *redis.Pool

func storeItem(key string, value string) {
	conn := pool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	processFatalError(err)
}

func getItem(key string) string {
	conn := pool.Get()
	defer conn.Close()

	val, err := redis.String(conn.Do("GET", key))

	if err != nil {
		return ""
	}

	return val
}

func createRedisConnection(server string, password string, db int) {

	pool = &redis.Pool{
		MaxIdle: 100,
		MaxActive: 10000,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", server, redis.DialDatabase(db), redis.DialPassword(password))
			processFatalError(err)
			return conn, err
		},
	}

	conn := pool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))

	processFatalError(err)

	if err == nil {
		log.Println("redis connection successful")
	}
}