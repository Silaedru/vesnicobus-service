package main

import (
	"github.com/gomodule/redigo/redis"
	"log"
)

var pool *redis.Pool

func closeConnection(c redis.Conn) {
	err := c.Close()

	if err != nil {
		log.Println("WARNING! error when closing redis connection:", err)
	}
}

func storeItem(key string, value string) {
	conn := pool.Get()
	defer closeConnection(conn)

	_, err := conn.Do("SET", key, value)

	if err != nil {
		log.Printf("WARNING! error when storing redis key \"%s\": %v\n", key, err)
	}
}

func getItem(key string) string {
	conn := pool.Get()
	defer closeConnection(conn)

	val, err := redis.String(conn.Do("GET", key))

	if err != nil {
		return ""
	}

	return val
}

func createRedisConnection(server string, password string, db int, maxIdle int, maxActive int) {
	pool = &redis.Pool{
		MaxIdle:   maxIdle,
		MaxActive: maxActive,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", server, redis.DialDatabase(db), redis.DialPassword(password))
			processFatalError(err)
			return conn, err
		},
	}

	conn := pool.Get()
	defer closeConnection(conn)

	_, err := redis.String(conn.Do("PING"))

	processFatalError(err)

	if err == nil {
		log.Println("redis connection successful")
	}
}
