package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

// setup HTML templates
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

// render home page
func Home(w http.ResponseWriter, r *http.Request) {
	if err := renderPage(w, "home.jet", nil); err != nil {
		log.Println(err)
	}
}

// load and render HTML template
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

// setup websocket upgrader
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// define the response sent back from websocket
type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

// cache websocket connection
type WebSocketConection struct {
	*websocket.Conn
}

// define the payload sent back to client via websocket
type WsPayload struct {
	Action   string             `json:"action"`
	UserName string             `json:"username"`
	Message  string             `json:"message"`
	Conn     WebSocketConection `json:"-"`
}

var (
	wsChan  = make(chan WsPayload)                // receive messages from client
	clients = make(map[WebSocketConection]string) // websocket connection: username
)

// upgrade connection to websocket
func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client connected to websocket")

	var response WsJsonResponse
	response.Action = "enter"
	response.Message = "Connected to server"
	response.ConnectedUsers = getUserList()

	// cache current client's websocket connection
	conn := WebSocketConection{
		Conn: ws,
	}
	clients[conn] = ""

	// format response to json and send it to only current client
	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	// run goroutine that listens for client's websocket
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
		err := conn.ReadJSON(&payload) // read message sent from client
		if err != nil {
			// do nothing
			break // "Error: repeated read on failed websocket connection" not logged
		} else {
			payload.Conn = *conn
			wsChan <- payload // push decoded message to the channel
		}
	}
}

func ListenToWSChannel() {
	var response WsJsonResponse

	// pop messages from channel
	for e := range wsChan {

		switch e.Action {
		case "username":
			// get a list of all users and send it back via broadcast
			clients[e.Conn] = e.UserName
			response.Action = "list_users"
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)

		case "left":
			// remove client data from map and refresh user list of other clients
			response.Action = "list_users"
			delete(clients, e.Conn)
			response.ConnectedUsers = getUserList()
			broadcastToAll(response)

		case "broadcast":
			// broadcast message to all connected clients
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.UserName, e.Message)
			broadcastToAll(response)
		}
	}
}

// send response to all connected clients
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

// get user list of connected clients
func getUserList() (userList []string) {
	for _, user := range clients {
		if user != "" {
			userList = append(userList, user)
		}
	}
	sort.Strings(userList)
	return userList
}
