package handlers

import (
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

func Home(w http.ResponseWriter, r *http.Request) {
	if err := renderPage(w, "home.jet", nil); err != nil {
		log.Println(err)
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	if err = view.Execute(w, data, nil); err != nil {
		return err
	}

	return nil
}

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// define the response sent back from websocket
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json"message"`
	MessageType    string   `json"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WebSocketConection struct {
	*websocket.Conn
}

type WsPayload struct {
	Action   string             `json:"action"`
	UserName string             `json:"username"`
	Message  string             `json"message"`
	Conn     WebSocketConection `json:"-"`
}

var (
	wsChan  = make(chan WsPayload)
	clients = make(map[WebSocketConection]string)
)

// upgrade connection to websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client connected to websocket")

	var response WsJsonResponse
	response.Message = `<em><small>Connected to server</small></em>`

	conn := WebSocketConection{
		Conn: ws,
	}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWS(&conn)
}

func ListenForWS(conn *WebSocketConection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error:", r)
		}
	}()

	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// do nothing
			break // "Error: repeated read on failed websocket connection" not logged
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func ListenToWSChannel() {
	var response WsJsonResponse

	for e := range wsChan {

		switch e.Action {
		case "username":
			// get a list of all users and send it back via broadcast
			clients[e.Conn] = e.UserName
			response.Action = "list_users"
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)

		case "left":
			response.Action = "list_users"
			delete(clients, e.Conn)
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)
		}
	}
}

func broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket error: client is not connected")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

func getUserList() (userList []string) {
	for _, user := range clients {
		if user != "" {
			userList = append(userList, user)
		}
	}
	sort.Strings(userList)
	return userList
}
