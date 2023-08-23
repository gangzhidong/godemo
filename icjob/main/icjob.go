package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"github.com/gorilla/websocket"
)

var addr = ":8090"

var upgrader = websocket.Upgrader{} // use default options

func apiSub(w http.ResponseWriter, r *http.Request) {
	//处理请求
	qvals := r.URL.Query()
	log.Println("success call icjob ")
	cmdStr := qvals.Get("cmd")
	log.Println("cmdStr: ", cmdStr)
	fmt.Fprint(w, "i am icjob")

	//websocket连接icmanager
	u := url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/ws/icsub", RawQuery: "jobId=job1"}
	log.Printf("connecting to %s", u.String())
	cjob, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	defer cjob.Close()
	if err != nil {
		log.Fatalf("err: %+v\n", err)
		return
	}

	//执行命令
	go func() {
		cmd := exec.Command(cmdStr)
		cmd.Run()
		log.Println("exec cmd over.")
	}()

	//接收消息
	go func() {
		for {
			_, message, err := cjob.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	//发送定时消息
	ticker := time.NewTicker(time.Second)
	tickerOver := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	defer tickerOver.Stop()
	for {
		select {
		case t := <-ticker.C:
			err := cjob.WriteMessage(websocket.BinaryMessage, []byte("icjob:"+t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-tickerOver.C:
			log.Println("interrupt")
			err := cjob.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			return
		}
	}

}
func wsSub(w http.ResponseWriter, r *http.Request) {
	qvals := r.URL.Query()
	log.Print("qvals:", qvals)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("err:%v", err)
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	log.SetFlags(0)
	http.HandleFunc("/api/bsub", apiSub)
	http.HandleFunc("/ws/bsub", wsSub)
	log.Fatal(http.ListenAndServe(addr, nil))
}
