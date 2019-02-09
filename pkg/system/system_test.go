package system

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestHandleError(t *testing.T) {
	refError := errors.New("an error")
	var out bytes.Buffer
	log.SetOutput(&out)
	defer func() {
		r := recover()
		require.NotNil(t, r)
		require.EqualValues(t, fmt.Sprintf("%s", r), refError.Error())
	}()
	HandleError(refError)

	// Must not reach this.
	t.Fail()
}

func TestNewViperConfig(t *testing.T) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config_test")
	err := viper.ReadInConfig()
	require.Nil(t, err)

	c := NewViperConfig()
	require.Equal(t, "127.0.0.1:8083", c.WebServerAddress)
	require.Equal(t, 300, c.TS3RegisterTimer)
}

func TestNew(t *testing.T) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config_test")
	err := viper.ReadInConfig()
	require.Nil(t, err)

	sigChan := make(chan os.Signal, 1)
	sys := New(sigChan)
	require.Equal(t, "127.0.0.1:8083", sys.Config.WebServerAddress)
	require.Equal(t, 300, sys.Config.TS3RegisterTimer)
	require.Equal(t, sigChan, sys.SigChan)
}
