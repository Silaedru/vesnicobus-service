package main

import (
	"log"
	"sync"
	"time"
)

var golemioApiKey string

func processFatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.Println("===== VESNICOBUS =====")
	log.Println("Vesnicobus service is starting ...")
	s := loadSettings("vesnicobus.ini")
	golemioApiKey = s.Golemio.ApiKey

	createRedisConnection(s.Redis.Server, s.Redis.Password, s.Redis.DB, s.Redis.MaxIdle, s.Redis.MaxActive)
	setMicrosoftApiKey(s.Microsoft.ApiKey)

	if s.Webservice.RefreshInterval > 0 {
		log.Printf("refresh interval is %ds", s.Webservice.RefreshInterval)
		go func() {
			for {
				refreshBusesLastPosition()
				time.Sleep(time.Duration(s.Webservice.RefreshInterval) * time.Second)
			}
		}()
	}

	go func() {
		for {
			time.Sleep(24 * time.Hour)
			tripLockMapMutex.Lock()
			tripLocks = make(map[string]*sync.Mutex)
			tripLockMapMutex.Unlock()
		}
	}()

	setupWebService(s.Webservice.Bind)
}
