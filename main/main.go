package main

import (
	"crypto/sha1"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)


var (
	host = "192.168.1.202"
	port = "22"
	user = "root"
	pass = "root"
	password = "root"
	termlog = "./test_termlog"

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
    // Create sshlib.Connect
    con := &sshlib.Connect{
        // If you use x11 forwarding, please set to true.
        ForwardX11: true,

        // If you use ssh-agent forwarding, please set to true.
        // And after, run `con.ConnectSshAgent()`.
        ForwardAgent: false,
    }

    // Create ssh.AuthMethod
    authMethod := sshlib.CreateAuthMethodPassword(password)

    // If you use ssh-agent forwarding, uncomment it.
    // con.ConnectSshAgent()

    // Connect ssh server
    err := con.CreateClient(host, port, user, []ssh.AuthMethod{authMethod})
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    // Set terminal log
    con.SetLog(termlog, false)
    // Create Session
    session, err := con.CreateSession()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
	con.X11Forward(session)
    // // Start ssh shell
    con.Shell(session)
	
	session.CombinedOutput("xclock")
}