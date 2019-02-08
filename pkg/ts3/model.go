package ts3

// User defines a model for a database and represents a ts3 user.
type User struct {
	ID int `storm:"id,unique,increment" db:"id"`

	EveCharID     int32  `db:"eve_char_id"`
	EveCharName   string `db:"eve_char_name"`
	EveCorpTicker string `db:"eve_corp_ticker"`
	EveAlliTicker string `db:"eve_alli_ticker"`

	TS3UID    string `db:"ts3_uid"`
	TS3CLDBID string `db:"ts3_cldbid"`

	Active bool `db:"active"`
}

// Store defines an interface of how to interact with user model on db level.
type Store interface {
	Init()
	Drop()
	CreateUser(u *User)
	Users() []*User
	ActiveUsersCharIDs() []int32
	UpdateUser(u *User)
	SetUserInactiveByUID(uid string)
	TS3UIDExists(uid string) bool
}

// Service defines an interface of how to ineract with ts3 service.
type Service interface {
	Start()
	Stop()
	GetStore() Store
	ValidateUsers()
	CreateRegisterRecord(u *User)
}
