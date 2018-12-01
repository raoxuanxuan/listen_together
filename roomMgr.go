package main

type RoomMgr struct {
	rooms []*Room
}

func (rm *RoomMgr) run() {
	for {
		for _, room := range rm.rooms {
			room.connMgr.run()
		}
	}
}
