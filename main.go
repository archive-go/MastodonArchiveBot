package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	// "github.com/MakeGolangGreat/telegraph-go"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var cookie string
var secWebSocketProtocol string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	readConfig()

	addr := "wss://alive.bar/api/v1/streaming/?stream=public:local"
	log.Printf("connecting to %s", addr)

	header := http.Header{}
	header.Add("Cookie", cookie)
	header.Add("Host", "alive.bar")
	header.Add("Origin", "https://alive.bar")
	header.Add("Sec-WebSocket-Protocol", secWebSocketProtocol)

	c, _, err := websocket.DefaultDialer.Dial(addr, header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	defer close(done)

	exitHandler()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)

		// page := &telegraph.Page{
		// 	AccessToken: "b968da509bb76866c35425099bc0989a5ec3b32997d55286c657e6994bbb",
		// 	AuthorURL:   link,
		// 	Title:       "内容备份",
		// 	Data:        `<blockquote>本页面由开源程序「<a href="https://github.com/MakeGolangGreat/ArchiveBot">ArchiveBot</a>」生成，内容来自网络。</blockquote>`,
		// }

	}

}

func exitHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		os.Exit(0)
	}()
}

func errHandler(msg string, err error) {
	if err != nil {
		fmt.Printf("%s - %s\n", msg, err)
		os.Exit(1)
	}
}

// 从配置中读取配置
func readConfig() {
	file, err := os.OpenFile("./config.json", os.O_RDWR|os.O_CREATE, 0766) //打开或创建文件，设置默认权限
	errHandler("读取配置失败", err)
	defer file.Close()

	var conf config
	err2 := json.NewDecoder(file).Decode(&conf)
	errHandler("解码配置失败", err2)

	cookie = conf.Cookie
	secWebSocketProtocol = conf.SecWebSocketProtocol
	color.Green("读取配置成功")
}
