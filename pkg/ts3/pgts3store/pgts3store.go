package pgts3store

import (
	"github.com/jmoiron/sqlx"
	"github.com/prusya/eve-ts3-service/pkg/system"
	"github.com/prusya/eve-ts3-service/pkg/ts3"
)

const (
	storeName            = "pgts3store"
	createUserTableQuery = `
	CREATE TABLE IF NOT EXISTS "user"
	(
		id              SERIAL PRIMARY KEY,
		eve_char_id     INTEGER NOT NULL,
		eve_char_name   VARCHAR(50) NOT NULL,
		eve_corp_ticker VARCHAR(50) NOT NULL,
		eve_alli_ticker VARCHAR(50) NOT NULL,
		ts3_uid         VARCHAR(50) NOT NULL UNIQUE,
		ts3_cldbid      VARCHAR(50) NOT NULL UNIQUE,
		active          BOOLEAN
	)`
	createUserQuery = `
	INSERT INTO "user"
	(eve_char_id, eve_char_name, eve_corp_ticker, eve_alli_ticker, ts3_uid, ts3_cldbid, active)
	VALUES ($1, $2, $3, $4, $5, $6, $7)`
	setUserInactiveByUIDQuery = `
	UPDATE "user"
	SET active = 'f'
	WHERE ts3_uid = $1`
	updateUserQuery = `
	UPDATE "user"
	SET eve_char_id = $1,
		eve_char_name = $2,
		eve_corp_ticker = $3,
		eve_alli_ticker = $4,
		ts3_uid = $5,
		ts3_cldbid = $6,
		active = $7
	WHERE id = $8`
)

// Store implements ts3.Store interface backed by postgresql and sqlx.
type Store struct {
	db *sqlx.DB
}

// New creates a new Store.
func New(db *sqlx.DB) *Store {
	s := Store{
		db: db,
	}
	s.Init()

	return &s
}

// Init prepares db for usage.
func (s *Store) Init() {
	_, err := s.db.Exec(createUserTableQuery)
	system.HandleError(err, storeName+".Init")
}

// Drop placeholder.
func (s *Store) Drop() {}

// CreateUser stores a ts3.User record.
func (s *Store) CreateUser(u *ts3.User) {
	_, err := s.db.Exec(createUserQuery, u.EveCharID, u.EveCharName,
		u.EveCorpTicker, u.EveAlliTicker, u.TS3UID, u.TS3CLDBID, u.Active)
	system.HandleError(err, storeName+".CreateUser", u)
}

// Users returns all ts3.User records.
func (s *Store) Users() []*ts3.User {
	var users []*ts3.User
	err := s.db.Select(&users, `SELECT * FROM "user"`)
	system.HandleError(err, storeName+".Users")

	return users
}

// ActiveUsersCharIDs returns EveCharIDs of users with `Active` set to true.
func (s *Store) ActiveUsersCharIDs() []int32 {
	var ids []int32
	err := s.db.Select(&ids, `SELECT eve_char_id FROM "user" WHERE active`)
	system.HandleError(err, storeName+".ActiveUsersCharIDs")

	return ids
}

// TS3UIDExists checks if record with provided uid exists.
func (s *Store) TS3UIDExists(uid string) bool {
	var exists bool
	err := s.db.Get(&exists,
		`SELECT EXISTS(SELECT 1 FROM "user" WHERE ts3_uid=$1)`, uid)
	system.HandleError(err, storeName+".TS3UIDExists", "uid="+uid)

	return exists
}

// UpdateUser updates a ts3.User record.
func (s *Store) UpdateUser(u *ts3.User) {
	_, err := s.db.Exec(updateUserQuery, u.EveCharID, u.EveCharName,
		u.EveCorpTicker, u.EveAlliTicker, u.TS3UID, u.TS3CLDBID, u.Active, u.ID)
	system.HandleError(err, storeName+".UpdateUser", u)
}

// SetUserInactiveByUID sets `active` to false for provided uid.
func (s *Store) SetUserInactiveByUID(uid string) {
	_, err := s.db.Exec(setUserInactiveByUIDQuery, uid)
	system.HandleError(err, storeName+".SetUserInactiveByUID", "uid="+uid)
}
