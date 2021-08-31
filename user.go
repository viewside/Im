package main

import (
	"fmt"
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C chan string
	conn net.Conn
	server *Server
}

//NewUser 创建一个用户的API
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name : userAddr,
		Addr : userAddr,
		C : make(chan string),
		conn: conn,
		server: server,

	}
	go user.ListenMessage()

	return user

}

//用户上线的业务
func (t *User) Online() {
	t.server.mapLock.Lock()

	t.server.OnlineMap[t.Name] = t
	t.server.mapLock.Unlock()
	//广播当前用户上线消息
	t.server.BroadCast(t, "已上线\n")
}

func (t *User) Offline()  {

	//用户下线就从在线表中删除
	t.server.mapLock.Lock()
	delete(t.server.OnlineMap,t.Name)
	t.server.mapLock.Unlock()
	//广播当前用户上线消息
	t.server.BroadCast(t, "已下线\n")
}

func (t *User)sendMsg(msg string)  {
	t.conn.Write([]byte(msg))
}

func (t *User) doMessage(msg string)  {
	if msg == "who" {
		for _,user := range t.server.OnlineMap{
			onlineMsg := "[" + user.Addr +"]" + user.Name + ":" + "[online]\n"
			//t.sendMsg(onlineMsg)
			//也可以直接放进channel?
			t.C <- onlineMsg
		}

	}else if len(msg)> 7 && msg[:7] == "rename|"{
		newName := strings.Split(msg, "|")[1]
		_,ok := t.server.OnlineMap[newName]

		if ok{
			t.sendMsg(fmt.Sprintf("已经存在：%s \n",newName))
		}else {
			t.rename(newName)

		}


	}else if len(msg)>4 && msg[:3] == "to|" &&len(strings.Split(msg, "|")) == 3 {

		splitMsg := strings.Split(msg, "|")
		t.PrivateChat(splitMsg[1], splitMsg[2])

	}else {

		t.server.BroadCast(t,msg+"\n")
	}

}

func (t *User)rename(name string)  {
	t.server.mapLock.Lock()
	delete(t.server.OnlineMap,t.Name )
	t.Name = name
	t.server.OnlineMap[t.Name] = t

	t.server.mapLock.Unlock()
	t.sendMsg("修改成功," + "用户名已经更新为：" + t.Name + "\n")

}

func (t *User) PrivateChat(targetUserName, msg string )  {

	if targetUserName == ""{
		t.sendMsg("消息格式不正确，请使用\"to|张三|你好 \"格式。\n")
		return
	}

	targetUser, ok := t.server.OnlineMap[targetUserName]

	if !ok {
		t.sendMsg("用户不存在\n")
		return
	}

	if msg == ""{
		t.sendMsg("无消息内容，请重新发送。\n")
		return
	}

	targetUser.sendMsg(t.Name + "对你说：" + msg + "\n")

}
func (t *User) ListenMessage()  {
	for true {
		msg := <-t.C


		t.sendMsg(msg)
	}
}