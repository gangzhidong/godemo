package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var addr = "0.0.0.0:8080"

var upgrader = websocket.Upgrader{} // use default options
var connCache = sync.Map{}
var disconn = make(chan bool)
var job_ID = "job1"

func apiSub(w http.ResponseWriter, r *http.Request) {
	//接收bsub请求
	fmt.Fprintf(w, "success to icmanager \n")
	qvals := r.URL.Query()
	cmdStr := qvals.Get("cmd")
	log.Print("query cmd :", cmdStr)
	go func(cmd string) {
		//调度任务
		log.Println("start scheduler:")
		<-time.After(3 * time.Second)
		log.Println("end scheduler:")

		//调用icjob接口
		client := http.Client{Timeout: 5 * time.Second}
		urlStr := fmt.Sprintf("http://%s%s", "127.0.0.1:8090", "/api/bsub?cmd="+cmd)
		body := "icmanager call icjob"
		req, err := http.NewRequest(http.MethodPost, urlStr, bytes.NewReader([]byte(body)))
		res, err := client.Do(req)
		if err != nil {
			return
		}
		defer res.Body.Close()
		bytes, err := io.ReadAll(res.Body)
		bstr := string(bytes)
		log.Printf("%s \n", bstr)
	}(cmdStr)

}

func wsSub(w http.ResponseWriter, r *http.Request) {
	//处理请求
	qvals := r.URL.Query()
	jobId := qvals.Get("jobId")
	log.Println("icbusb websocket conn jobid:", jobId)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	//缓存链接
	connCache.Store(jobId, c)
	defer connCache.Delete(jobId)
	defer c.Close()

	//监听结束信号
	<-disconn
}

func icSub(w http.ResponseWriter, r *http.Request) {

	//处理请求
	qvals := r.URL.Query()
	jobId := qvals.Get("jobId")
	log.Println("icjob websocket conn jobid:", jobId)
	icjob, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer icjob.Close()

	//获取bsub的websocket链接
	connIntr, ok := connCache.Load(jobId)
	if !ok {
		return
	}
	icbsub := connIntr.(*websocket.Conn)

	//连接icjob和icbsub
	var wg sync.WaitGroup
	wg.Add(2)
	//转发icjob的数据到icbsub
	go func() {
		defer wg.Done()
		for {
			mt, message, err := icjob.ReadMessage()
			if err != nil {
				log.Println("from icjob read:", err)
				break
			}
			log.Printf("recv: %s", message)
			err = icbsub.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}()
	//转发icbsub的数据到icjob
	go func() {
		defer wg.Done()
		for {
			mt, message, err := icbsub.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("from icbsub recv: %s", message)
			err = icjob.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}()
	wg.Wait()
	//发送结束信号
	disconn <- true
}

func main() {
	log.SetFlags(0)
	http.HandleFunc("/ws/bsub", wsSub)
	http.HandleFunc("/api/bsub", apiSub)
	http.HandleFunc("/ws/icsub", icSub)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func IcmanagerInit() {
	main()
}
