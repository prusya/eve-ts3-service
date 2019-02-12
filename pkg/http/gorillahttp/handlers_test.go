package gorillahttp

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/prusya/eve-ts3-service/pkg/system"
	"github.com/prusya/eve-ts3-service/pkg/ts3/darfkts3service"
	"github.com/prusya/eve-ts3-service/pkg/ts3/pgts3store"
	"github.com/stretchr/testify/require"
)

func TestDeserializeEveChar(t *testing.T) {
	referenceEC := eveChar{
		EveCharID:     1,
		EveCorpID:     2,
		EveAlliID:     3,
		EveCharName:   "char name",
		EveCorpName:   "corp name",
		EveAlliName:   "alli name",
		EveCorpTicker: "corp ticker",
		EveAlliTicker: "alli ticker",
	}
	j, err := json.Marshal(&referenceEC)
	require.Nil(t, err)
	data := base64.StdEncoding.EncodeToString(j)

	ec := deserializeEveChar(data)
	require.Equal(t, referenceEC.EveCharID, ec.EveCharID)
	require.Equal(t, referenceEC.EveCharName, ec.EveCharName)
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
	referenceEC := eveChar{
		EveCharID:     1,
		EveCorpID:     2,
		EveAlliID:     3,
		EveCharName:   "char name",
		EveCorpName:   "corp name",
		EveAlliName:   "alli name",
		EveCorpTicker: "corp ticker",
		EveAlliTicker: "alli ticker",
	}
	j, err := json.Marshal(&referenceEC)
	require.Nil(t, err)
	data := base64.StdEncoding.EncodeToString(j)
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
