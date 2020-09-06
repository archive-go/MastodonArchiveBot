package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"regexp"

	"github.com/MakeGolangGreat/archive-go"
	"github.com/MakeGolangGreat/mastodon-go"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var cookie string
var mastodonToken string
var telegraphToken string
var secWebSocketProtocol string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	readConfig()
	listen()
}

func listen() {
	addr := "wss://alive.bar/api/v1/streaming/?stream=public:local"

	header := http.Header{}
	header.Add("Cookie", cookie)
	header.Add("Host", "alive.bar")
	header.Add("Origin", "https://alive.bar")
	header.Add("Sec-WebSocket-Protocol", secWebSocketProtocol)

	ws, _, err := websocket.DefaultDialer.Dial(addr, header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()
	exitHandler()

	for {
		_, body, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		var message stream
		if err := json.Unmarshal(body, &message); err != nil {
			color.Red("解析字符串出错！", err)
		}

		switch message.Event {
		case "update":
			var status mastodon.Status
			err := json.Unmarshal([]byte(message.Payload.(string)), &status)
			if err != nil {
				fmt.Println("err", err)
			}
			// 不监测自己的发嘟
			if status.Account.UserName == "beifen" {
				continue
			}

			linkRegExp := regexp.MustCompile(`href="(http.*?)"`)

			// 如果监测到链接存在，交给archive-go备份
			if linkRegExp.MatchString(status.Content) {
				matchURL := linkRegExp.FindAllSubmatch([]byte(status.Content), -1)

				var totalURL string
				for _, url := range matchURL {
					link := string(url[1])

					archivelink, saveError := archive.Save(link, telegraphToken, attachInfo)
					if saveError != nil {
						fmt.Println("文章保存出错：", saveError.Error())
					} else {
						totalURL += archivelink + "\n"
					}
				}

				fmt.Printf("备份链接：%s，长度%d\n", totalURL, len(totalURL))

				if len(totalURL) == 0 {
					fmt.Println("无备份链接生成（可能是出错，也可能是链接都已经备份过）")
					continue
				}

				reply := fmt.Sprintf("监测到链接存在...\n\n备份链接内容到 Telegraph 成功...\n\n输出链接：\n%s\n\n#备份\n\n本Bot代码开源：%s", totalURL, projectLink)

				toot := &mastodon.Mastodon{
					Token:  mastodonToken,
					Domain: "https://alive.bar",
				}
				result, err := toot.SendStatuses(&mastodon.StatusParams{
					Status:      reply,
					MediaIds:    "[]",
					Poll:        "[]",
					Visibility:  "public",
					InReplyToID: status.ID,
					Sensitive:   true,
					SpoilerText: "自动备份",
				})
				if err != nil {
					color.Red(err.Error())
					continue
				}

				fmt.Printf("回嘟成功，ID：%s\n", result.ID)
			}
		}
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
	mastodonToken = conf.MastodonToken
	telegraphToken = conf.TelegraphToken
	color.Green("读取配置成功")
}
