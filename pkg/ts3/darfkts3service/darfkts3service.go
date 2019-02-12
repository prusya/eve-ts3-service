package darfkts3service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	client "github.com/darfk/ts3"
	"github.com/pkg/errors"

	"github.com/prusya/eve-ts3-service/pkg/system"
	"github.com/prusya/eve-ts3-service/pkg/ts3"
)

const (
	serviceName = "darfkts3service"
)

// Service implements ts3.Service interface backed by darfk/ts3 lib.
type Service struct {
	system    *system.System
	client    *client.Client
	store     ts3.Store
	registerQ map[string]registerRecord
	lock      sync.RWMutex
	stopChan  chan struct{}
}

// userData defines a response from users validation server.
type userData struct {
	EveCharID     int32
	EveCorpTicker string
	EveAlliTicker string
	Valid         bool
}

type registerRecord struct {
	at   int64
	user *ts3.User
}

// New creates a new service and prepares it to start.
func New(system *system.System, store ts3.Store) *Service {
	s := Service{
		system:    system,
		store:     store,
		registerQ: make(map[string]registerRecord),
		stopChan:  make(chan struct{}, 1),
	}

	s.system.TS3 = &s

	return &s
}

// Start starts the service.
func (s *Service) Start() {
	// Connect to ts3 server.
	c, err := client.NewClient(s.system.Config.TS3Address)
	system.HandleError(err)
	// Login.
	_, err = c.Exec(client.Login(s.system.Config.TS3User,
		s.system.Config.TS3Password))
	system.HandleError(err)
	// Select virtual server.
	_, err = c.Exec(client.Use(s.system.Config.TS3ServerID))
	system.HandleError(err)
	// Subscribe to server notifications to receive messages about new
	// connections.
	_, err = c.Exec(client.Command{
		Command: "servernotifyregister",
		Params: map[string][]string{
			"event": []string{"server"},
		},
	})
	system.HandleError(err)

	s.client = c
	s.client.NotifyHandler(s.eventHandler)

	keepAliveT := time.NewTicker(60 * time.Second)
	rqCleanupT := time.NewTicker(300 * time.Second)
	validateUsersT := time.NewTicker(20 * time.Minute)
	go func() {
		select {
		case <-keepAliveT.C:
			go s.keepAlive()
		case <-rqCleanupT.C:
			go s.registerQCleanup()
		case <-validateUsersT.C:
			go s.ValidateUsers()
		case <-s.stopChan:
			keepAliveT.Stop()
			rqCleanupT.Stop()
			validateUsersT.Stop()
			break
		}
	}()
}

// Stop stops the Service.
func (s *Service) Stop() {
	s.stopChan <- struct{}{}
	if s.client != nil {
		s.client.ExecString("quit")
		s.client.Close()
	}
}

// GetStore returns ts3.Store.
func (s *Service) GetStore() ts3.Store {
	return s.store
}

func recoverPanic() {
	recover()
}

// ValidateUsers keeps user records up to date, assigns proper ts3 server goups,
// deletes users from ts3 server if they don't have access to ts3 service.
func (s *Service) ValidateUsers() {
	defer recoverPanic()

	// We need to check only active users.
	ids := s.store.ActiveUsersCharIDs()

	// Send ids to the validation server.
	payload, _ := json.Marshal(ids)
	resp, err := http.Post(s.system.Config.UsersValidationEndpoint,
		"application/json", bytes.NewBuffer(payload))
	system.HandleError(err, serviceName+".ValidateUsers http.Post", ids)
	if resp.StatusCode != 200 {
		err = errors.New("non 200 response from validation server")
		system.HandleError(err, serviceName+".ValidateUsers", resp.StatusCode)
	}

	usersData := []userData{}
	d := json.NewDecoder(resp.Body)
	d.UseNumber()
	d.Decode(&usersData)
	resp.Body.Close()

	// Process response from the validation server.
	users := s.store.Users()
	for _, u := range usersData {
		for _, user := range users {
			if user.EveCharID == u.EveCharID {
				// Invalid users should only be removed from server groups.
				if !u.Valid {
					s.allServerGroupsDelClient(user.TS3CLDBID)
					s.store.SetUserInactiveByUID(user.TS3UID)
					continue
				}

				// Move user to a proper group if corp or alli has changed.
				if user.EveCorpTicker != u.EveCorpTicker ||
					user.EveAlliTicker != u.EveAlliTicker {
					// Remove user from current group.
					currentGroup := fmt.Sprintf("%s %s", user.EveAlliTicker,
						user.EveCorpTicker)
					found, sgid := s.serverGroupByName(currentGroup)
					if found {
						s.serverGroupDelClient(sgid, user.TS3CLDBID)
					}

					// Add user to a new group.
					newGroup := fmt.Sprintf("%s %s", u.EveAlliTicker,
						u.EveCorpTicker)
					found, sgid = s.serverGroupByName(newGroup)
					if !found {
						sgid = s.serverGroupCopy(newGroup)
					}
					s.serverGroupAddClient(sgid, user.TS3CLDBID)

					// Finally, update store record.
					user.EveCorpTicker = u.EveCorpTicker
					user.EveAlliTicker = u.EveAlliTicker
					s.store.UpdateUser(user)
				}
			}
		}
	}
}

// allServerGroupsDelClient removes user from all server groups
// except `server admin`.
func (s *Service) allServerGroupsDelClient(cldbid string) {
	resp, err := s.client.Exec(client.Command{
		Command: "servergroupsbyclientid",
		Params: map[string][]string{
			"cldbid": []string{cldbid},
		},
	})
	system.HandleError(err, serviceName+".deleteTS3User")

	for _, group := range resp.Params {
		// Skip `server admin` group.
		if group["sgid"] == "6" {
			continue
		}
		s.serverGroupDelClient(group["sgid"], cldbid)
	}
}

// serverGroupDelClient removes user from a server group.
func (s *Service) serverGroupDelClient(sgid, cldbid string) {
	_, err := s.client.Exec(client.Command{
		Command: "servergroupdelclient",
		Params: map[string][]string{
			"sgid":   []string{sgid},
			"cldbid": []string{cldbid},
		},
	})
	system.HandleError(err, serviceName+".serverGroupDelClient",
		"sgid="+sgid, "cldbid="+cldbid)
}

// serverGroupAddClient adds user to a server group.
func (s *Service) serverGroupAddClient(sgid, cldbid string) {
	_, err := s.client.Exec(client.Command{
		Command: "servergroupaddclient",
		Params: map[string][]string{
			"sgid":   []string{sgid},
			"cldbid": []string{cldbid},
		},
	})
	system.HandleError(err, serviceName+".serverGroupAddClient",
		"sgid="+sgid, "cldbid="+cldbid)
}

// serverGroupCopy creates a new group by copying the reference group.
func (s *Service) serverGroupCopy(groupName string) string {
	resp, err := s.client.Exec(client.Command{
		Command: "servergroupcopy",
		Params: map[string][]string{
			"ssgid": []string{s.system.Config.TS3ReferenceGroupID},
			"tsgid": []string{"0"},
			"type":  []string{"1"},
			"name":  []string{groupName},
		},
	})
	system.HandleError(err, serviceName+".serverGroupCopy", "groupName="+groupName)

	sgid, ok := resp.Params[0]["sgid"]
	if !ok {
		err := errors.New("missing sgid in response")
		system.HandleError(err, serviceName+".serverGroupCopy",
			"groupName="+groupName, resp.Params)
	}

	return sgid
}

// serverGroupByName returns whether server group exists and its sgid.
func (s *Service) serverGroupByName(groupName string) (bool, string) {
	resp, err := s.client.Exec(client.Command{
		Command: "servergrouplist",
	})
	system.HandleError(err, serviceName+".serverGroupByName", "groupName="+groupName)

	for _, group := range resp.Params {
		if group["name"] == groupName {
			return true, group["sgid"]
		}
	}

	return false, "sgid"
}

// CreateRegisterRecord creates a new register record.
func (s *Service) CreateRegisterRecord(u *ts3.User) {
	s.lock.Lock()
	s.registerQ[u.EveCharName] = registerRecord{
		at:   time.Now().Unix(),
		user: u,
	}
	s.lock.Unlock()
}

// eventHandler receives server events.
// It is responsible for adding users to a proper server group.
func (s *Service) eventHandler(n client.Notification) {
	defer recoverPanic()

	// We need only `notifycliententerview` event with `reasonid=0`.
	// This event occurs when a user connects to the server.
	if n.Type != "notifycliententerview" {
		return
	}
	if n.Params[0]["reasonid"] != "0" {
		return
	}

	cluid := n.Params[0]["client_unique_identifier"]
	clnickname := n.Params[0]["client_nickname"]
	cldbid := n.Params[0]["client_database_id"]

	// Check if connected user is in the register queue.
	record, ok := s.registerQ[clnickname]
	if !ok {
		return
	}
	// Check if the registration record didn't expire.
	now := time.Now().Unix()
	if record.at+int64(s.system.Config.TS3RegisterTimer) < now {
		return
	}

	// Add user to the proper group.
	groupName := fmt.Sprintf("%s %s", record.user.EveAlliTicker,
		record.user.EveCorpTicker)
	found, sgid := s.serverGroupByName(groupName)
	// If group not found, create it.
	if !found {
		sgid = s.serverGroupCopy(groupName)
	}
	s.serverGroupAddClient(sgid, cldbid)

	// Finally, store user record in the db.
	record.user.TS3CLDBID = cldbid
	record.user.TS3UID = cluid
	record.user.Active = true
	if s.store.TS3UIDExists(cluid) {
		s.store.UpdateUserByUID(record.user)
	} else {
		s.store.CreateUser(record.user)
	}
}

// keepAlive is actually a `version` command.
// Use this func perioducally to keep connection alive.
func (s *Service) keepAlive() {
	defer recoverPanic()

	_, err := s.client.Exec(client.Version())
	system.HandleError(err, serviceName+".keepAlive")
}

// registerQCleanup removes expired register records.
func (s *Service) registerQCleanup() {
	now := time.Now().Unix()
	s.lock.Lock()
	for k, v := range s.registerQ {
		if v.at+int64(s.system.Config.TS3RegisterTimer) < now {
			delete(s.registerQ, k)
		}
	}
	s.lock.Unlock()
}
