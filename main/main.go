package main

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)


var (
	host = "192.168.1.202"
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
	conn, err := net.Dial("unix", os.Getenv("DISPLAY"))
	if err != nil {
		return
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		io.Copy(conn, channel)
		conn.(*net.UnixConn).CloseWrite()
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
	// x11-req指定有效负载
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

	// 将按键输入转换为连接目标可以识别的格式（这里是按键）
	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		fmt.Println(err)
	}
	defer terminal.Restore(fd, state)

	// 获取终端尺寸
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

	// Shell启动
	err = session.Shell()
	if err != nil {
		fmt.Println(err)
	}

	// 终端尺寸变化检测和处理
	signal_chan := make(chan os.Signal, 1)
	signal.Notify(signal_chan, syscall.SIGWINCH)
	// signal.Notify(signal_chan, syscall.SIGTERM)
	go func() {
		for {
			s := <-signal_chan
			switch s {
			case syscall.SIGWINCH:
			// case syscall.SIGTERM:
				fd := int(os.Stdout.Fd())
				w, h, _ = terminal.GetSize(fd)
				session.WindowChange(h, w)
			}
		}
	}()
	session.Run("xclock")
	err = session.Wait()
	if err != nil {
		fmt.Println(err)
	}
}