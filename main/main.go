package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

var (
	host = "192.168.1.202"
	port = "22"
	user = "root"
	pass = "root"
	password = "root"
	termlog = "./test_termlog"

)

func main() {
	// SSH连接配置
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// SSH连接
	client, err := ssh.Dial("tcp", host+":"+port, config)
	if err != nil {
		log.Fatalf("Failed to dial: %s", err)
	}
	defer client.Close()

	// 创建SSH会话
	session, err := client.NewSession()
	if err != nil {
		log.Fatalf("Failed to create session: %s", err)
	}
	defer session.Close()

	// X11转发
	x11Conn, err := net.Dial("unix", os.Getenv("DISPLAY"))
	if err != nil {
		log.Fatalf("Failed to connect to X11 server: %s", err)
	}
	defer x11Conn.Close()

	// 将X11连接绑定到SSH会话
	session.Stdin = x11Conn
	session.Stdout = x11Conn
	session.Stderr = os.Stderr

	// 执行X11程序
	err = session.Run("xclock")
	if err != nil {
		log.Fatalf("Failed to run X11 program: %s", err)
	}

	fmt.Println("X11 program executed successfully")
}