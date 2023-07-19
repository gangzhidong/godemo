package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"

	"crypto/sha1"
	"errors"
	"math/rand"
	"time"

	"golang.org/x/crypto/ssh"
)

var (
	host           = "192.168.1.202"
	port           = "22"
	user           = "root"
	pass           = "root"
	password       = "root"
	termlog        = "./test_termlog"
	characterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

type x11request struct {
	SingleConnection bool
	AuthProtocol     string
	AuthCookie       string
	ScreenNumber     uint32
}

// forwardX11Socket ssh.Channel forward socket
func forwardX11Socket(channel ssh.Channel, session *ssh.Session) {
	conn, err := net.Dial("unix", "/tmp/.X11-unix/X3")
	if err != nil {
		fmt.Printf("err:%v \n", err)
		return
	}
	// session.Stdin = conn
	session.Stdout = conn
	session.Stderr = os.Stderr
	err =session.Run("xclock")
	if err != nil {
		fmt.Printf("err:%v \n", err)
		return
	}
	// var wg sync.WaitGroup
	// wg.Add(2)
	// go func() {
	// 	io.Copy(conn, channel)
	// 	conn.(*net.UnixConn).CloseWrite()
	// 	wg.Done()
	// }()
	// go func() {
	// 	io.Copy(channel, conn)
	// 	channel.CloseWrite()
	// 	wg.Done()
	// }()

	// wg.Wait()
	// conn.Close()
	// channel.Close()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewSHA1Hash generates a new SHA1 hash based on
// a random number of characters.
func NewSHA1Hash(n ...int) string {
	noRandomCharacters := 32

	if len(n) > 0 {
		noRandomCharacters = n[0]
	}

	randString := RandomString(noRandomCharacters)

	hash := sha1.New()
	hash.Write([]byte(randString))
	bs := hash.Sum(nil)

	return fmt.Sprintf("%x", bs)
}

// RandomString generates a random string of n length
func RandomString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = characterRunes[rand.Intn(len(characterRunes))]
	}
	return string(b)
}

func main() {
	
	// Create sshClientConfig
	sshConfig := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// SSH connect.
	client, err := ssh.Dial("tcp", host+":"+port, sshConfig)

	// Create Session
	session, err := client.NewSession()
	defer session.Close()

	// NOTE:
	// x11-reqのPayloadを指定
	payload := x11request{
		SingleConnection: false,
		AuthProtocol:     string("MIT-MAGIC-COOKIE-1"),
		AuthCookie:       string(NewSHA1Hash()),
		ScreenNumber:     uint32(0),
	}

	// Send x11-req Request
	ok, err := session.SendRequest("x11-req", true, ssh.Marshal(payload))
	if err == nil && !ok {
		fmt.Println(errors.New("ssh: x11-req failed"))
	} else {
		// Open HandleChannel x11
		x11channels := client.HandleChannelOpen("x11")
		conn, err := net.Dial("unix", "/tmp/.X11-unix/X3")
		if err != nil {
			fmt.Printf("err:%v \n", err)
			return
		}
		var outbuf bytes.Buffer
		var errbuf bytes.Buffer
		session.Stdin = conn
		session.Stdout = &outbuf
		session.Stderr = &errbuf

		go func ()  {
			for{
				<-time.Tick(time.Second)
				fmt.Printf("outbuf:%v \n", outbuf)
				fmt.Printf("errtbuf:%v \n", errbuf)
			}
			
		}() 
		go session.Run("xclock")
		for ch := range x11channels {
			channel, _, err := ch.Accept()
			if err != nil {
				continue
			}

			go forwardX11Socket(channel, session)
		}
	}

	// キー入力を接続先が認識できる形式に変換する(ここがキモ)

}

func maintest() {
	_, err1 := net.Dial("unix", "/tmp/.X11-unix/X"+os.Args[1])
	if err1 != nil {
		log.Fatalf("err1: %s", err1)
	}

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
	x11Conn, err := net.Dial("unix", "/tmp/.X11-unix/X0")
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
