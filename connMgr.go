package main

import "fmt"

/*
* ConnMgr用来管理同一个房间里的用户连接
 */
type ConnMgr struct {
	users      map[*User]bool
	register   chan *User
	unregister chan *User
	broadcast  chan []byte
}

func NewConnMgr() *ConnMgr {
	return &ConnMgr{
		register:   make(chan *User),
		unregister: make(chan *User),
		broadcast:  make(chan []byte),
		users:      make(map[*User]bool),
	}
}

func (mgr *ConnMgr) run() {

	for {
		select {
		case user := <-mgr.register:
			//fmt.Println(user.uin, "user <- mgr.register")
			//fmt.Printf("add user addr %p\n", user)
			mgr.users[user] = true
		case user := <-mgr.unregister:
			//fmt.Println(user.uin, "user <- mgr.unregister")
			if _, ok := mgr.users[user]; ok {
				delete(mgr.users, user)
				close(user.send)
			}
		case msg := <-mgr.broadcast:
			//fmt.Println(string(msg), "msg <- mgr.broadcast")
			for user := range mgr.users {
				select {
				case user.send <- msg:
					fmt.Println("connMgr: ", user.uin, " user.send <- msg ", string(msg))
				default:
					fmt.Println("should close user, uin:", user.uin)
					//close(user.send)
					//delete(mgr.users, user)
				}
			}
		}
	}
}
