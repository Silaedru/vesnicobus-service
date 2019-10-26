package main

import (
	"gopkg.in/gcfg.v1"
)

type Settings struct {
	Webservice struct {
		Bind string
		RefreshInterval int
	}

	Redis struct {
		Server string
		Password string
		DB int
	}

	Golemio struct {
		ApiKey string
	}

	Microsoft struct {
		ApiKey string
	}
}

func loadSettings(path string) Settings {
	var s Settings
	err := gcfg.ReadFileInto(&s, path)
	processFatalError(err)
	return s
}