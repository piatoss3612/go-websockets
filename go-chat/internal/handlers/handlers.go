package handlers

import (
	"fmt"
	"log"
	"net/http"

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
	Action      string `json:"action"`
	Message     string `json"message"`
	MessageType string `json"message_type"`
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
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func ListenToWSChannel() {
	var response WsJsonResponse

	for e := range wsChan {
		response.Action = "Got here"
		response.Message = fmt.Sprintf("Some message, and action was %s", e.Action)
		broadcastToAll(response)
	}
}

func broadcastToAll(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket error: client might not be connected")
			_ = client.Close()
			delete(clients, client)
		}
	}
}
