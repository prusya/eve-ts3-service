package darfkts3service

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/prusya/eve-ts3-service/pkg/system"
	"github.com/prusya/eve-ts3-service/pkg/ts3"
	"github.com/prusya/eve-ts3-service/pkg/ts3/pgts3store"
)

func TestDarfkts3service(t *testing.T) {
	sys := &system.System{
		Config: &system.Config{
			TS3RegisterTimer: 300,
		},
	}
	db := &sqlx.DB{}
	store := pgts3store.New(db)

	t.Run("TestNew", func(t *testing.T) {
		ts3service := New(sys, store)
		require.Equal(t, ts3service, sys.TS3)
	})

	t.Run("TestGetStore", func(t *testing.T) {
		ts3service := New(sys, store)
		require.Equal(t, store, ts3service.GetStore())
	})

	t.Run("TestCreateRegisterRecord", func(t *testing.T) {
		ts3service := New(sys, store)
		user := &ts3.User{
			EveCharName: "test user",
		}
		ts3service.CreateRegisterRecord(user)
		record, ok := ts3service.registerQ[user.EveCharName]
		require.True(t, ok)
		require.Equal(t, user, record.user)
	})

	t.Run("TestRegisterQCleanup", func(t *testing.T) {
		ts3service := New(sys, store)
		ts3service.lock.Lock()
		ts3service.registerQ["test user 2"] = registerRecord{
			at: 1,
		}
		ts3service.lock.Unlock()
		ts3service.registerQCleanup()
		_, ok := ts3service.registerQ["test user 2"]
		require.False(t, ok)
	})
}
