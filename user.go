package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// 对客户端展示的用户信息
type UserInfo struct {
	Uin    int64  `json:"uin"`
	Avatar string `json:"avatr"`
	Online bool   `json:"online"`
}

// 内部使用的用户信息结构体
type User struct {
	uin           int64
	room_ids      []int64
	last_fav_time int32
	conn          *websocket.Conn
	send          chan []byte
	room          *Room // 用户当前所在的房间
}

func NewUser(conn *websocket.Conn) *User {
	return &User{
		send: make(chan []byte),
		conn: conn,
	}
}
func (user *User) set_uin(uin int64) {
	user.uin = uin
}
func (user *User) set_room_ids(room_ids []int64) {
	user.room_ids = room_ids[:]
}

// read from websocket and push it to mgr
func (user *User) read() {
	fmt.Println("user read start")
	defer func() {
		fmt.Println("user read end")
		if user.room != nil {
			user.room.connMgr.unregister <- user
			close(user.send)
			user.conn.Close()
		}
	}()

	for {
		mt, msg, err := user.conn.ReadMessage()
		if err != nil {
			fmt.Printf("user.conn.ReadMessage failed!err:%v", err)
			break
		}
		var param WsParams
		json_err := json.Unmarshal(msg, &param)
		if json_err != nil {
			fmt.Println(json_err)
			return
		}
		fmt.Println("user.uin:", user.uin, " GetRequest:", string(msg))
		switch param.Ws_action_type {
		//case kCreateRoom:
		//case kJoinRoom:
		case kExitRoom:
			user.room.connMgr.unregister <- user
			/*
				case kDestroyRoom:
					fmt.Println("destory room")
				case kPreSong:
					fmt.Println("pre song")
					user.room.connMgr.broadcast <- msg
				case kNextSong:
					fmt.Println("next song")
					user.room.connMgr.broadcast <- msg
				case kAddSong:
					fmt.Println("add song")
					user.room.connMgr.broadcast <- msg
				case kDeleteSong:
					fmt.Println("delete song")
					user.room.connMgr.broadcast <- msg
				case kFav:
					fmt.Println("fav song")
					user.room.connMgr.broadcast <- msg
			*/
		default:
			user.room.connMgr.broadcast <- msg
		}
		rsp := &WsResponse{
			Result: 0,
			Msg:    "ok",
		}
		send_response(mt, user.conn, &rsp)
	}
}

// read from mgr and push it to websocket
func (user *User) write() {
	ticker := time.NewTicker(time.Second * 2)
	fmt.Println("user write start")
	defer func() {
		fmt.Println("user write end")
		if user.room != nil {
			user.room.connMgr.unregister <- user
			close(user.send)
			user.conn.Close()
		}
	}()

	for {
		select {
		case <-ticker.C:
			ping := &struct {
				Text string
			}{
				Text: "ping",
			}
			msg, _ := json.Marshal(ping)
			if err := user.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				fmt.Println("send ping message failed!", err)
			} else {
				fmt.Println("send ping message")
			}

		case msg, ok := <-user.send:
			if !ok {
				user.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			ws_params := &WsParams{}
			json.Unmarshal(msg, ws_params)
			if ws_params.Uin == user.uin {
				// 不处理自己发出的请求
				fmt.Println(string(msg), "ws_params.uin == user.uin  ", user.uin)
				continue
			}
			fmt.Println(string(msg), "msg <- user.send  ", user.uin)
			switch ws_params.Ws_action_type {
			case kJoinRoom:
				// 如果是joinRoom请求，需要告诉对端ready
				/*
					r_p := &WsParams{
						Ws_action_type: kQueryPlaylistID,
						Uin:            ws_params.Uin,
					}

					new_msg, err := json.Marshal(r_p)
					if err != nil {
						fmt.Println(err)
					}
					user.room.connMgr.broadcast <- new_msg
				*/
			default:
			}
			/*
				w, err := user.conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
			*/
			uin_str := strconv.FormatInt(user.uin, 10)
			var buffer bytes.Buffer
			buffer.WriteString("uin ")
			buffer.WriteString(uin_str)
			buffer.WriteString(" receive ")
			buffer.WriteString(string(msg))
			//w.Write([]byte(buffer.String()))
			err := user.conn.WriteMessage(websocket.TextMessage, []byte(buffer.String()))
			if err != nil {
				fmt.Println("writemessage err:", err)
				return
			}
			fmt.Println(user.uin, " send to websocket:", buffer.String())
		}
	}
}

type WsRsp struct {
}

func (user *User) run() {
	go user.read()
	go user.write()
}
func send_response(msg_type int, conn *websocket.Conn, rsp interface{}) {
	send_data, err := json.Marshal(rsp)
	if err != nil {
		fmt.Println(err)
	} else {
		conn.WriteMessage(msg_type, send_data)
		fmt.Println("send_data:", string(send_data))
	}
}
