package handlers

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/pusher/pusher-http-go"
)

// authenticate the user to pusher server
func (repo *DBRepo) PusherAuth(w http.ResponseWriter, r *http.Request) {
	userID := repo.App.Session.GetInt(r.Context(), "userID")

	u, err := repo.DB.GetUserById(userID)
	if err != nil {
		log.Println(err)
		return
	}

	params, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		return
	}

	presenceData := pusher.MemberData{
		UserID: strconv.Itoa(userID),
		UserInfo: map[string]string{
			"name": u.FirstName,
			"id":   strconv.Itoa(userID),
		},
	}

	response, err := app.WsClient.AuthenticatePresenceChannel(params, presenceData)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(response)
	if err != nil {
		log.Println(err)
		return
	}
}

func (repo *DBRepo) TestPusher(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]string)
	data["message"] = "Hello, World"

	err := repo.App.WsClient.Trigger("public-channel", "test-event", data)
	if err != nil {
		log.Println(err)
		return
	}
}
