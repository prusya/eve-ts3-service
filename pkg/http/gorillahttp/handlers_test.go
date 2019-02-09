package gorillahttp

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/prusya/eve-ts3-service/pkg/system"
	"github.com/prusya/eve-ts3-service/pkg/ts3/darfkts3service"
	"github.com/prusya/eve-ts3-service/pkg/ts3/pgts3store"
	"github.com/stretchr/testify/require"
)

func TestDeserializeUser(t *testing.T) {
	formatted := fmt.Sprintf("charName=%s;charID=%d;corpName=%s;corpID=%d;"+
		"corpTicker=%s;alliName=%s;alliID=%d;alliTicker=%s;",
		"char name", 1, "corp name", 2, "corp ticker", "alli name", 3, "alli ticker")
	data := base64.StdEncoding.EncodeToString([]byte(formatted))
	user := deserializeUser(data)
	require.Equal(t, "char name", user.EveCharName)
	require.Equal(t, "corp ticker", user.EveCorpTicker)
	require.Equal(t, "alli ticker", user.EveAlliTicker)
	require.Equal(t, int32(1), user.EveCharID)
}

func TestCreateRegisterRecord(t *testing.T) {
	sys := &system.System{
		Config: &system.Config{
			WebServerAddress: ":8081",
			TS3RegisterTimer: 300,
		},
	}
	db := &sqlx.DB{}
	store := pgts3store.New(db)
	darfkts3service.New(sys, store)
	httpservice := New(sys)

	httpservice.Start()
	formatted := fmt.Sprintf("charName=%s;charID=%d;corpName=%s;corpID=%d;"+
		"corpTicker=%s;alliName=%s;alliID=%d;alliTicker=%s;",
		"char name", 1, "corp name", 2, "corp ticker", "alli name", 3, "alli ticker")
	data := base64.StdEncoding.EncodeToString([]byte(formatted))
	req, _ := http.NewRequest("GET", "http://localhost:8081/api/ts3/v1/createregisterrecord", nil)
	req.AddCookie(&http.Cookie{
		Name:  "char",
		Value: data,
	})
	resp, err := http.DefaultClient.Do(req)
	require.Nil(t, err)
	require.Equal(t, 200, resp.StatusCode)
	resp.Body.Close()
	httpservice.Stop()
}
