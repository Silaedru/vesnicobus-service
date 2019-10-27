package main

import (
	"log"
	"time"
)

var golemioApiKey string

func processFatalError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	defer log.Println("Vesnicobus service has been stopped.")

	log.Println("===== VESNICOBUS =====")
	log.Println("Vesnicobus service is starting ...")
	s := loadSettings("vesnicobus.ini")
	golemioApiKey = s.Golemio.ApiKey

	createRedisConnection(s.Redis.Server, s.Redis.Password, s.Redis.DB)
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

	setupWebService(s.Webservice.Bind)
}
