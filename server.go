package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type  Server struct {
	Ip string
	Port int
	//在线用户的列表
	OnlineMap map[string]*User
	mapLock sync.RWMutex
	//消息广播的channel
	Message chan string
}
//创建一个server接口
func NewServer(ip string,port int) *Server{
	server:=&Server{
		Ip: ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
	return server
}
//监听Message广播channel的grountine，一旦有消息就发送给全部的在线User
func (this *Server)ListenMessager()  {
	for  {
		msg:=<-this.Message
		//将msg发送给全部的在线user
		this.mapLock.Lock()
		for _,cli:=range  this.OnlineMap{
			cli.C<-msg
		}
		this.mapLock.Unlock()
	}
}
//广播消息的方法
func (this *Server)BroadCast(user *User,msg string)  {
	senMsg:="["+user.Addr+"]"+user.Name+":"+msg
	this.Message<-senMsg
}
func (this *Server) Handler(conn net.Conn)  {
	//// .. 当前连接的业务
	//fmt.Println("连接成功")
	user:=NewUser(conn)
	//用户上线，将用户加入onlinemap表中
	this.mapLock.Lock()
	this.OnlineMap[user.Name]=user
	this.mapLock.Unlock()
	//广播当前用户消息
	this.BroadCast(user,"以上线")
	//接受客户端发送的消息
	go func() {
		buf:=make([]byte,4096)
		for  {
			n,err:=conn.Read(buf)
			if n==0 {
				this.BroadCast(user,"下线")
				return
			}
			if err!=nil &&err!=io.EOF{
				fmt.Println("conn read err:",err)
				return
			}
			//提取用户的消息（去除\n)
			msg:=string(buf[:n-1])
			//将得到的消息进行广播
			this.BroadCast(user,msg)
		}
	}()
	//当前handle阻塞
	select{}
}
func (this *Server) Start() {
	//socket listen
listener,err:=net.Listen("tcp",fmt.Sprintf("%s:%d",this.Ip,this.Port))
if err!=nil{
	fmt.Println("net.listen is err",err)
	return
}
	//close listen socket
	defer listener.Close()
	//启动监听Message的groutine
	go this.ListenMessager()
for {
	//accept
conn,err:=listener.Accept()
if err!=nil{
	fmt.Println("listerner accpet is err",err)
	continue
}
	//do handler
	go this.Handler(conn)
}

}
