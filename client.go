package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP string
	ServerPort int
	Name string
	conn net.Conn
	flaG int
}

func NewClient(serverIP string, serverPort int)  *Client{

	client := &Client{
		ServerIP: serverIP,
		ServerPort: serverPort,
		flaG: 99,
	}
	//建立连接
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil{
		fmt.Println("net.Dial.error", err)
		return nil
	}

	client.conn = conn

	return client
}

//处理server 回应的消息，直接显示在标准输出
func (client *Client) DealResponse() {
	io.Copy(os.Stdout, client.conn)
	/*
	io.copy相当于
	for{
		buf := make([]byte, len)
		n,err = conn.Read(buf)
		if err != nil{...}
		fmt.println(buf)
		}
	 */
}
func (client *Client) menu() bool {

	var flaG int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	_, err := fmt.Scanln(&flaG)

	if err != nil {
		println("fmt.Scanln err:", err)
	}

	if flaG >= 0 && flaG <= 3  {
		client.flaG = flaG
		return true
	}else {
		fmt.Println(">>>>>请输入合法范围的数字<<<<<")
		return false
	}

}

func (client *Client) UpdateName() bool {

	fmt.Println(">>>>>请输入用户名：")
	var name string
	fmt.Scanln(&name)

	sendMsg := "rename|" + name +"\n"

	_, err := client.conn.Write([]byte(sendMsg))

	if err != nil{
		println("conn.Write err:", err)
		return false
	}

	client.Name = name
	return true

}

func (client *Client) PublicChat()  {
	var chatMsg string

	for {
		fmt.Println(">>>>>请输入聊天内容,exit退出")
		fmt.Scanln(&chatMsg)
		if chatMsg == "exit"{
			break
		}
		if len(chatMsg) !=0 {
			_, err := client.conn.Write([]byte(chatMsg + "\n"))
			if err != nil {
				fmt.Println("conn.Write err:", err)
				return
			}
		}
	}
}
func (client *Client)selectOnlineUser()  {
	_, err :=  client.conn.Write([]byte("who\n"))
	if err != nil{
		fmt.Println("conn.write err:", err)
		return
	}
}

func (client *Client)PrivateChat()  {
	var targetUserName string
	var chatMsg string
	client.selectOnlineUser()

	for{
		fmt.Println("请输入要私聊的对象[用户名],exit退出")
		fmt.Scanln(&targetUserName)
		if targetUserName == "exit" {
			return
		} else if targetUserName == "" {
			fmt.Println("用户名不能为空")
		} else {
			break
		}
	}

	for{
		println("正在和[" + targetUserName + "]" + "聊天" + "请输入聊天内容，exit退出。")
		fmt.Scanln(&chatMsg)
		if chatMsg == "exit"{
			return
		}
		sendMsg := fmt.Sprintf("to|%s|%s\n", targetUserName, chatMsg)
		_, err := client.conn.Write([]byte(sendMsg))

		if err != nil{
			fmt.Println("conn.Write err:", err)
			return
		}
	}





}

func (client *Client) run(){

	for client.flaG != 0{
		for !client.menu(){}
		switch client.flaG {
		case 1:
			fmt.Println("公聊模式选择...")
			client.PublicChat()

		case 2:
			fmt.Println("私聊模式选择...")
			client.PrivateChat()

		case 3:
			fmt.Println("改名模式选择...")
			client.UpdateName()



		}
	}
}


/*
以命令行的方式解析输入参数
./client -ip 127.0.0.1 -port 88888
 */
var serverIP string
var serverPort int

func init()  {
	flag.StringVar(&serverIP,"ip", "127.0.0.1", "设置服务器的IP地址（默认是127.0.0.0.1）")
	flag.IntVar(&serverPort,"port", 8888, "设置服务器的IP地址（默认是8888）")
}

func main()  {
	//解析
	flag. Parse()
	client := NewClient(serverIP, serverPort)
	if client == nil{
		fmt.Println(">>>>>连接服务器失败...")
		return
	}
	//开启一个goroutine 处理服务器的响应
	go client.DealResponse()

	fmt.Println(">>>>>连接服务器成功...")
	client.run()





}

