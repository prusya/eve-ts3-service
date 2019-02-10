package gorillahttp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/prusya/eve-ts3-service/pkg/system"
	"github.com/prusya/eve-ts3-service/pkg/ts3"
)

var (
	reCharID     = regexp.MustCompile(".*?charID=([^;]+);.*|.*")
	reCorpID     = regexp.MustCompile(".*?corpID=([^;]+);.*|.*")
	reAlliID     = regexp.MustCompile(".*?alliID=([^;]+);.*|.*")
	reCharName   = regexp.MustCompile(".*?charName=([^;]+);.*|.*")
	reCorpName   = regexp.MustCompile(".*?corpName=([^;]+);.*|.*")
	reAlliName   = regexp.MustCompile(".*?alliName=([^;]+);.*|.*")
	reCorpTicker = regexp.MustCompile(".*?corpTicker=([^;]+);.*|.*")
	reAlliTicker = regexp.MustCompile(".*?alliTicker=([^;]+);.*|.*")
)

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

// HealthcheckH calls respondOK.
func HealthcheckH(w http.ResponseWriter, r *http.Request) {
	respondOK(w)
}

// CreateRegisterRecordH creates a new register record for ts3 service.
func (s *Service) CreateRegisterRecordH(w http.ResponseWriter, r *http.Request) {
	defer recoverPanic(w)

	cookie, err := r.Cookie("char")
	system.HandleError(err, serviceName+".CreateRegisterRecordHandler")
	user := deserializeEveUser(cookie.Value)
	s.system.TS3.CreateRegisterRecord(user)

	respondWithJSON(w, 200, s.system.Config.TS3RegisterTimer)
}

func deserializeEveUser(data string) *ts3.User {
	bytes, err := base64.StdEncoding.DecodeString(data)
	system.HandleError(err, serviceName+".deserializeEveUser", "data="+data)
	decoded := string(bytes)

	user := ts3.User{
		EveCharName:   reCharName.ReplaceAllString(decoded, "$1"),
		EveCorpTicker: reCorpTicker.ReplaceAllString(decoded, "$1"),
		EveAlliTicker: reAlliTicker.ReplaceAllString(decoded, "$1"),
	}
	charIDstr := reCharID.ReplaceAllString(decoded, "$1")
	charID64, err := strconv.ParseInt(charIDstr, 10, 32)
	system.HandleError(err, serviceName+".deserializeEveUser", "data="+data)
	user.EveCharID = int32(charID64)

	return &user
}
