package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)


func main() {
	//请求http
	addr := "127.0.0.1:8080"
	body := "hello I am icsub."
	client := http.Client{Timeout: 5 * time.Second}
	urlStr := fmt.Sprintf("http://%s%s", addr, "/api/bsub?cmd=date")
	req, err := http.NewRequest(http.MethodPost, urlStr, bytes.NewReader([]byte(body)))
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	bstr := string(bytes)
	log.Printf("resp: %s \n", bstr)

	//请求websocket
	u := url.URL{Scheme: "ws", Host: addr, Path:  "/ws/bsub", RawQuery: "jobId=job1"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	//读取数据
	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	//发送数据
	ticker := time.NewTicker(time.Second)
	tickerOver := time.NewTicker(60*time.Second)
	defer ticker.Stop()
	defer tickerOver.Stop()
	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.BinaryMessage, []byte("icsub:"+t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-tickerOver.C:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			return
		}
	}
}