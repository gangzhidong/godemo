package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

var (
	host = "node01"
	port = "22"
	user = "root"
	pass = "root"

	characterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

type x11request struct {
	SingleConnection bool
	AuthProtocol     string
	AuthCookie       string
	ScreenNumber     uint32
}

// forwardX11Socket ssh.Channel forward socket
func forwardX11Socket(channel ssh.Channel) {
	// conn, err := net.Dial("tcp", "localhost:6014")
	conn, err := net.Dial("unix", os.Getenv("DISPLAY"))
	if err != nil {
		return
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(conn, channel)
		conn.(*net.TCPConn).CloseWrite()
		wg.Done()
	}()
	go func() {
		io.Copy(channel, conn)
		channel.CloseWrite()
		wg.Done()
	}()

	wg.Wait()
	conn.Close()
	channel.Close()
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

		go func() {
			for ch := range x11channels {
				channel, _, err := ch.Accept()
				if err != nil {
					continue
				}

				go forwardX11Socket(channel)
			}
		}()
	}

	// キー入力を接続先が認識できる形式に変換する(ここがキモ)
	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		fmt.Println(err)
	}
	defer terminal.Restore(fd, state)

	// ターミナルサイズの取得
	w, h, err := terminal.GetSize(fd)
	if err != nil {
		fmt.Println(err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	err = session.RequestPty("xterm", h, w, modes)
	if err != nil {
		fmt.Println(err)
	}

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	// Shellを起動
	err = session.Shell()
	if err != nil {
		fmt.Println(err)
	}

	// ターミナルサイズの変更検知・処理
	// signal_chan := make(chan os.Signal, 1)
	// signal.Notify(signal_chan, syscall.SIGWINCH)
	// go func() {
	// 	for {
	// 		s := <-signal_chan
	// 		switch s {
	// 		case syscall.SIGWINCH:
	// 			fd := int(os.Stdout.Fd())
	// 			w, h, _ = terminal.GetSize(fd)
	// 			session.WindowChange(h, w)
	// 		}
	// 	}
	// }()

	err = session.Wait()
	if err != nil {
		fmt.Println(err)
	}
}