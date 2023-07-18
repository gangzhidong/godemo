package ssh_test

import (
	"fmt"
	"os"
	"testing"

	sshlib "github.com/blacknon/go-sshlib"
	"golang.org/x/crypto/ssh"
)

var (
    host     = "192.168.1.202"
    port     = "22"
    user     = "root"
    password = "root"

    termlog = "./test_termlog"
)
func TestSimplyOut (t *testing.T) {
    fmt.Println("test hello")
}
func TestSshLib (t *testing.T) {
    // Create sshlib.Connect
    con := &sshlib.Connect{
        // If you use x11 forwarding, please set to true.
        ForwardX11: true,

        // If you use ssh-agent forwarding, please set to true.
        // And after, run `con.ConnectSshAgent()`.
        ForwardAgent: true,
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

    // Start ssh shell
    con.Shell(session)
	session.CombinedOutput("xclock")
}