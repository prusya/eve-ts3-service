package system

import (
	"log"
	"os"
	"runtime/debug"

	"github.com/spf13/viper"

	"github.com/prusya/eve-ts3-service/pkg/http"
	"github.com/prusya/eve-ts3-service/pkg/ts3"
)

// System contains objects sharable across packages.
type System struct {
	TS3     ts3.Service
	HTTP    http.Service
	Config  *Config
	SigChan chan os.Signal
}

// Config contains all configurable options.
type Config struct {
	WebServerAddress string

	TS3Address          string
	TS3User             string
	TS3Password         string
	TS3ServerID         int
	TS3Whitelisted      string
	TS3ReferenceGroupID string
	TS3RegisterTimer    int

	UsersValidationEndpoint string

	PgConnString string
}

// New creates a new System.
func New(sigChan chan os.Signal) *System {
	config := NewViperConfig()
	s := System{
		SigChan: sigChan,
		Config:  config,
	}

	return &s
}

// NewViperConfig creates a new Config populated with values by viper.
func NewViperConfig() *Config {
	var c Config
	err := viper.Unmarshal(&c)
	HandleError(err, "system.NewViperConfig")

	return &c
}

// HandleError logs the error and panics.
func HandleError(err error, params ...interface{}) {
	if err == nil {
		return
	}

	if len(params) > 0 {
		log.Println("provided data:")
		for _, p := range params {
			log.Printf("\n%#+v\n", p)
		}
	}
	log.Printf("\n%s\n%s\n\n", err, string(debug.Stack()))

	panic(err)
}
