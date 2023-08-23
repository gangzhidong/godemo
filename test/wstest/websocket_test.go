package wstest

import (
	"fmt"
	"log"
	"net/url"
	"testing"

	icmanager "github.com/gangzhidong/godemo/icmanager/main"

	"github.com/gorilla/websocket"
)


func TestIcmanager (t *testing.T) {
	icmanager.IcmanagerInit()
}

func TestSimplyOut (t *testing.T) {
	// dialer := websocket.Dialer{
	// 	NetDial: func(network, addr string) (net.Conn, error) {
	// 		gotAddr = addr
	// 		return net.Dial(network, addrs[tt.server])
	// 	},
	// 	TLSClientConfig: tls,
	// }
	// h := http.Header{}
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:8090", Path: "/api/bsub"}
	log.Printf("connecting to %s", u.String())
    cjob, _, err :=  websocket.DefaultDialer.Dial(u.String(), nil)
    // cjob, _, err := dialer.Dial("ws://localhost:8090/api/icbsub", h)
	if err == nil {
		fmt.Printf("err: %+v\n", err)
		cjob.Close()
	}
}