package gorillahttp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prusya/eve-ts3-service/pkg/system"
	"github.com/prusya/eve-ts3-service/pkg/ts3"
)

type eveChar struct {
	EveCharID     int32
	EveCorpID     int32
	EveAlliID     int32
	EveCharName   string
	EveCorpName   string
	EveAlliName   string
	EveCorpTicker string
	EveAlliTicker string
	Valid         bool
}

// respondWithJSON receives a payload of any type, converts it into json
// and writes resulting json to a response writer.
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError receives a message string, converts it into json
// and writes resulting json to a response writer.
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// respondOK writes json `status: ok` to w.
func respondOK(w http.ResponseWriter) {
	respondWithJSON(w, 200, map[string]string{"status": "ok"})
}

// respond401 responds with http Unauthorized 401.
func respond401(w http.ResponseWriter) {
	respondWithError(w, 401, http.StatusText(401))
}

// respond403 responds with http Forbidden 403.
func respond403(w http.ResponseWriter) {
	respondWithError(w, 403, http.StatusText(403))
}

// NotFoundH responds with http Not Found 404.
func NotFoundH(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 404, http.StatusText(404))
}

// MethodNotAllowedH responds with http Method Not Allowed 405.
func MethodNotAllowedH(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 405, http.StatusText(405))
}

// recoverPanic recovers and responds with http 500 in case of a panic.
func recoverPanic(w http.ResponseWriter) {
	if r := recover(); r != nil {
		respondWithError(w, 500, fmt.Sprintf("%s", r))
	}
}

// HealthCheckH calls respondOK.
func HealthCheckH(w http.ResponseWriter, r *http.Request) {
	respondOK(w)
}

// CreateRegisterRecordH creates a new register record for ts3 service.
func (s *Service) CreateRegisterRecordH(w http.ResponseWriter, r *http.Request) {
	defer recoverPanic(w)

	cookie, err := r.Cookie("char")
	system.HandleError(err, serviceName+".CreateRegisterRecordHandler")
	eu := deserializeEveChar(cookie.Value)
	user := ts3.User{
		EveCharID:     eu.EveCharID,
		EveCharName:   eu.EveCharName,
		EveCorpTicker: eu.EveCorpTicker,
		EveAlliTicker: eu.EveAlliTicker,
		Active:        true,
	}
	s.system.TS3.CreateRegisterRecord(&user)

	respondWithJSON(w, 200, s.system.Config.TS3RegisterTimer)
}

// deserializeEveChar converts base64 encoded json with eve char data into struct.
func deserializeEveChar(data string) *eveChar {
	// Decode base64 into json.
	j, err := base64.StdEncoding.DecodeString(data)
	system.HandleError(err, serviceName+".deserializeUser DecodeString", "data="+data)

	// Decode json into struct.
	var ec eveChar
	d := json.NewDecoder(bytes.NewReader(j))
	d.UseNumber()
	err = d.Decode(&ec)
	system.HandleError(err, serviceName+".deserializeEveChar Decode", "data="+data)

	return &ec
}
