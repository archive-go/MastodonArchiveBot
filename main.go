package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/MakeGolangGreat/archive-go"
	"github.com/MakeGolangGreat/mastodon-go"
	"github.com/PuerkitoBio/goquery"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var cookie string
var mastodonToken string
var telegraphToken string
var domain string

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	readConfig()
	listen()
	// createConfig()
}

func listen() {
	addr := "wss://" + domain + "/api/v1/streaming/?stream=public:local"

	header := http.Header{}
	header.Add("Cookie", cookie)
	header.Add("Host", domain)
	header.Add("Origin", "https://"+domain)

	ws, _, err := websocket.DefaultDialer.Dial(addr, header)
	if err != nil {
		log.Fatal("dial:", err.Error())
	}
	defer ws.Close()
	exitHandler()

	fmt.Println("连接上WS，持续监听")
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
				fmt.Println("err", err.Error())
			}
			// 不监测自己的发嘟
			if status.Account.UserName == "beifen" {
				continue
			}

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(status.Content))
			if err != nil {
				fmt.Println("goquery 解析HTML字符串失败：", err.Error())
			}

			var totalURL string
			doc.Find("a").Each(func(_ int, s *goquery.Selection) {
				href, exists := s.Attr("href")
				if !exists {
					return
				}

				url, err := url.Parse(href)
				if err != nil {
					fmt.Println("解析URL失败", err.Error())
					return
				}
				fmt.Println(url.Host)
				if url.Host == domain {
					// 如果链接是当前实例的，那么不会将其备份
					return
				}

				// 如果监测到链接存在，交给archive-go备份
				archivelink, saveError := archive.Save(href, telegraphToken, attachInfo, &archive.More{
					IncludeAll: false,
				})
				if saveError != nil {
					fmt.Println("文章保存出错：", saveError.Error())
				} else {
					totalURL += archivelink + "\n"
				}
			})

			if len(totalURL) == 0 {
				continue
			}

			reply := fmt.Sprintf("监测到链接存在...\n\n备份链接内容到 Telegraph 成功...\n\n输出链接：\n%s\n\n#备份\n\n本Bot代码开源：%s", totalURL, projectLink)

			toot := &mastodon.Mastodon{
				Token:  mastodonToken,
				Domain: "https://" + domain,
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
	domain = conf.Domain
	mastodonToken = conf.MastodonToken
	telegraphToken = conf.TelegraphToken
	color.Green("读取配置成功")
}

// 检测是否链接中包含一些敏感参数
// 比如微信公众号的链接包含了 sharer_shareid 参数
func leakSecretInfo(link string) (bool, string) {
	if url, err := url.Parse(link); err != nil {
		host := url.Hostname()
		q := url.Query()
		if host == "mp.weixin.qq.com" {
			return q.Get("sharer_shareid") != "", "警告⚠️微信文章链接中包含 ‘sharer_shareid’ 这样的隐私参数，它看上去可以追查到文章分享者。"
		} else if host == "music.163.com" {
			return q.Get("userid") != "", "警告⚠️网易云音乐链接中包含 ‘userid’参数，它可以非常容易地被用来反查到特定的用户。"
		}
	}
	return false, ""
}
