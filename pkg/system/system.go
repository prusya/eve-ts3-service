package system

import (
	"log"
	"os"
	"runtime/debug"

	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/prusya/eve-ts3-service/pkg/http"
	"github.com/prusya/eve-ts3-service/pkg/ts3"
)

type System struct {
	TS3     ts3.Service
	HTTP    http.Service
	Config  *Config
	SigChan chan os.Signal
}

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

func New(sigChan chan os.Signal) (*System, error) {
	config, err := NewViperConfig()
	if err != nil {
		err = errors.WithMessage(err, "System")
	}

	s := System{
		SigChan: sigChan,
		Config:  config,
	}

	return &s, err
}

func NewViperConfig() (*Config, error) {
	var err error
	var c Config

	err = viper.Unmarshal(&c)
	if err != nil {
		err = errors.WithMessage(err, "NewViperConfig: viper.Unmarshal")
	}

	return &c, err
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
