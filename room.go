package main

import (
	"fmt"

	"github.com/gorilla/websocket"
)

type RoomStatus int32

const (
	kRoomNotExist = 1
	kRoomFull     = 2
	kRoomOk       = 3
)

type CreateRoomReq struct {
	user  *User
	cover string
	name  string
	uin   int64
	conn  *websocket.Conn
}
type CreateRoomRsp struct {
	Result  int32 `json:"result"`
	Reason  int32 `json:"reason"`
	Room_id int64 `json:"room_id"`
}
type JoinRoomReq struct {
	uin     int64
	room_id int64
	user    *User
	msg     []byte
	conn    *websocket.Conn
}
type JoinRoomRsp struct {
	Result    int32     `json:"result"`
	Reason    int32     `json:"reason"`
	Room_info *RoomInfo `json:"room_info"`
}
type RoomInfo struct {
	Room_id int64       `json:"room_id"`
	Cover   string      `json:"cover"`
	Users   []*UserInfo `json:"users"`
	Name    string      `json:"name"`
}

type Room struct {
	room_id int64
	name    string
	cover   string
	connMgr *ConnMgr
}

var g_room_id int64 = 0

var rooms = make(map[int64]*Room)

func IncreaseRoomID() int64 {
	g_room_id++
	return g_room_id
}

func GetRoomInfo(room_id int64) *Room {
	if room, ok := rooms[room_id]; ok {
		return room
	} else {
		return nil
	}
}

func checkRoomStatus(room *Room) int32 {
	online_user_cnt := 0
	for _, v := range room.connMgr.users {
		if v == true {
			online_user_cnt++
		}
	}
	fmt.Printf("online_user_cnt:%d\n", online_user_cnt)
	if online_user_cnt == 2 {
		return kRoomFull
	} else {
		return kRoomOk
	}
}
func RecordRoom(room *Room) {
	rooms[room.room_id] = room
}

func CreateRoom(req *CreateRoomReq, rsp *CreateRoomRsp) int32 {

	rsp.Room_id = IncreaseRoomID()
	room := &Room{
		room_id: rsp.Room_id,
		cover:   req.cover,
		name:    req.name,
	}

	user := NewUser(req.conn)
	user.set_uin(req.uin)
	var room_ids []int64
	room_ids = append(room_ids, rsp.Room_id)
	user.set_room_ids(room_ids)
	room.connMgr = NewConnMgr()
	user.room = room

	go user.run()
	fmt.Println("go user.run()")

	go room.connMgr.run()
	fmt.Println("go room.connMgr.run()")

	room.connMgr.register <- user
	fmt.Println(user.uin, " room.connMgr.register <- user")

	RecordRoom(room)
	return 0
}

func JoinRoom(req *JoinRoomReq, rsp *JoinRoomRsp) int32 {

	room := GetRoomInfo(req.room_id)
	if room == nil {
		rsp.Result = kResultErr
		rsp.Reason = kRoomNotExist
		return 1
	} else {
		fmt.Println("room is not null, show room info:", *room)
		/*
			fmt.Println("id:", room.room_id)
			for u, _ := range room.connMgr.users {
				fmt.Println("user :", u.uin)
			}
		*/
	}
	room_status := checkRoomStatus(room)
	if room_status != kRoomOk {
		rsp.Result = kResultErr
		rsp.Reason = room_status
		return 1
	}
	if room.connMgr == nil {
		fmt.Println("room.connMgr is null, create room")
		room.connMgr = NewConnMgr()
		go room.connMgr.run()
	}
	user := NewUser(req.conn)
	user.set_uin(req.uin)
	var room_ids []int64
	room_ids = append(room_ids, room.room_id)
	user.set_room_ids(room_ids)
	user.room = room

	go user.run()
	room.connMgr.register <- user
	fmt.Println("room.connMgr.register <- req.user", user.uin)
	room.connMgr.broadcast <- req.msg
	fmt.Println("room.connMgr.broadcast <- req.msg", string(req.msg))

	room_info := &RoomInfo{
		Room_id: room.room_id,
		Cover:   room.cover,
		Name:    room.name,
	}
	rsp.Room_info = room_info
	for room_user, online := range room.connMgr.users {
		user_info := &UserInfo{
			Uin:    room_user.uin,
			Avatar: "",
			Online: online,
		}
		fmt.Println(*user_info)
		rsp.Room_info.Users = append(rsp.Room_info.Users, user_info)
	}
	return 0
}
