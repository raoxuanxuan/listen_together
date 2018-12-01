package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}
var roomMgr = RoomMgr{}

type ApiActionType int32
type WsActionType int32

const (
	kLogin ApiActionType = 1
)
const (
	kCreateRoom WsActionType = 1 + iota
	kJoinRoom
	kExitRoom
	kDestroyRoom
	kPreSong
	kNextSong
	kAddSong
	kDeleteSong
	kFav
	kQueryPlaylistID
	kQuerySongInfo
	kBuildPlaylist
)
const (
	kResultOK  = 0
	kResultErr = 1
)

type ApiParams struct {
	api_action_type ApiActionType `json:"action_type"`
	uin             int64         `json:"uin"`
}
type WsParams struct {
	Ws_action_type WsActionType `json:"ws_type"`
	Uin            int64        `json:"uin"`
	Room_id        int64        `json:"room_id"`
	Cover          string       `json:"cover"`
	Name           string       `json:"name"`
}
type WsResponse struct {
	Result int32  `json:"result"`
	Msg    string `json:"msg"`
}

func serveApi(w http.ResponseWriter, r *http.Request) {
	// api接口里处理login
	if r.Method != "POST" {
		http.Error(w, "please send a post request", 400)
		return
	}
	if r.Body == nil {
		http.Error(w, "body is empty", 400)
		return
	}
	var param ApiParams
	json.NewDecoder(r.Body).Decode(&param)
	fmt.Println(param)
	switch param.api_action_type {
	case kLogin:
		Login(param.uin)
	}
}
func Login(uin int64) {
	fmt.Println("Login")
	return
}
func serveHome(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "")
}

func parseParams(conn *websocket.Conn) {
	mt, param, err := conn.ReadMessage()
	if err != nil {
		return
	}
	var p WsParams
	json_err := json.Unmarshal(param, &p)
	if json_err != nil {
		fmt.Println(json_err)
		return
	}
	fmt.Println("GetRequest:", string(param))

	switch p.Ws_action_type {
	case kCreateRoom:
		req := &CreateRoomReq{
			uin:   p.Uin,
			name:  p.Name,
			cover: p.Cover,
			conn:  conn,
		}
		rsp := &CreateRoomRsp{}
		CreateRoom(req, rsp)
		send_response(mt, conn, rsp)

	case kJoinRoom:
		req := &JoinRoomReq{
			room_id: p.Room_id,
			uin:     p.Uin,
			msg:     param,
			conn:    conn,
		}
		rsp := &JoinRoomRsp{}
		JoinRoom(req, rsp)
		send_response(mt, conn, rsp)

	case kExitRoom:
	case kDestroyRoom:
	case kPreSong:
	case kNextSong:
	case kAddSong:
	case kDeleteSong:
	case kFav:
	default:
		fmt.Println("ws_action_type is invalid!")
	}

}
func serveWs(w http.ResponseWriter, r *http.Request) {
	// serverWs 开启新协程来处理读写请求
	/*
		if r.Method != "POST" {
			http.Error(w, "please send a post request", 400)
			fmt.Println("method is ", r.Method)
			return
		}
	*/
	if r.Body == nil {
		http.Error(w, "body is empty", 400)
		fmt.Println("body is empty")
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	parseParams(conn)
}

func main() {
	http.HandleFunc("/", serveHome)   // 展示测试页面
	http.HandleFunc("/api", serveApi) // 处理http请求
	http.HandleFunc("/ws", serveWs)   // 处理websocket请求
	err := http.ListenAndServe("localhost:10000", nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
	go roomMgr.run()
}

var homeTemplate = template.Must(template.New("").Parse(` khhhhhhh
<html>
<head>test</head>
<body>test</body>
</html>
`))
