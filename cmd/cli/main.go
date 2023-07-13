package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	logformat "github.com/antonfisher/nested-logrus-formatter"
	"github.com/jing332/tts-server-go/server"
	log "github.com/sirupsen/logrus"
)

var port = flag.Int64("port", 1233, "自定义监听端口")
var token = flag.String("token", "", "使用token验证")
var useDnsEdge = flag.Bool("use-dns-edge", false, "使用DNS解析Edge接口，而不是内置的北京微软云节点。")

func heartbeat() {
	url := "http://api.effectlib.com/v2/heartbeat"

	// 创建一个 HTTP 客户端
	client := &http.Client{
		Timeout: 5 * time.Second, // 设置超时时间
	}

	// 创建一个无限循环，定时发送心跳请求
	for {
		// 发送 GET 请求
		if resp, err := client.Get(url); err != nil {
			fmt.Println("心跳请求发送失败:", err)
		} else {
			// 检查响应状态码
			if resp.StatusCode != http.StatusOK {
				fmt.Println("心跳请求返回非 200 状态码:", resp.StatusCode)
			} else {
				fmt.Println("心跳正常")
			}
			defer resp.Body.Close()
		}
		// 处理心跳响应
		// ...

		// 休眠一段时间后再次发送心跳请求
		time.Sleep(10 * time.Second)
	}
}

func main() {
	go heartbeat()
	log.SetFormatter(&logformat.Formatter{HideKeys: true,
		TimestampFormat: "01-02|15:04:05",
	})
	flag.Parse()
	if *token != "" {
		log.Info("使用Token: ", token)
	}
	if *useDnsEdge == true {
		log.Infof("使用DNS解析Edge接口")
	}

	srv := &server.GracefulServer{Token: *token, UseDnsEdge: *useDnsEdge}
	srv.HandleFunc()

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		srv.Close()
	}()

	if err := srv.ListenAndServe(*port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server ListenAndServe: %v", err)
	}
	log.Infoln("服务已关闭")
}
