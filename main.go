package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"

	"github.com/MakeGolangGreat/archive-go"
	"github.com/MakeGolangGreat/mastodon-go"
	"github.com/PuerkitoBio/goquery"
	"github.com/ansel1/merry"
	"github.com/fatih/color"
	"github.com/gorilla/websocket"
)

var mastodonToken string
var telegraphToken string
var domain string

func main() {
	readConfig()
	listen()
}

func listen() {
	addr := "wss://" + domain + "/api/v1/streaming/?stream=public:local"
	ws, res, err := websocket.DefaultDialer.Dial(addr, nil)
	defer ws.Close()
	if err != nil {
		fmt.Println(res)
		merry.Wrap(err)
	}

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
				if url.Host == domain {
					// 如果链接是当前实例的，那么不会将其备份
					return
				}

				// 监测到链接中包含参数，私信提醒
				if hasLeak, msg := leakSecretInfo(href); hasLeak {
					go func(username string, _msg string) {
						toot := &mastodon.Mastodon{
							Token:  mastodonToken,
							Domain: "https://" + domain,
						}
						reminderResult, reminderErr := toot.SendStatuses(&mastodon.StatusParams{
							Status:      "@" + username + "\n\n" + _msg + " \n\n如果我以后不想收到此类提醒消息怎么办？\n好办，直接 Block 本Bot即可",
							MediaIds:    "[]",
							Poll:        "[]",
							Visibility:  "direct",
							InReplyToID: status.ID,
							Sensitive:   false,
						})
						if reminderErr != nil {
							color.Red(reminderErr.Error())
						}
						fmt.Println("私信成功", reminderResult.ID)
					}(status.Account.UserName, msg)
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

// 从配置中读取配置
func readConfig() {
	file, err := os.OpenFile("./config.json", os.O_RDWR|os.O_CREATE, 0766) //打开或创建文件，设置默认权限
	merry.Wrap(err)
	defer file.Close()

	var conf config
	err = json.NewDecoder(file).Decode(&conf)
	merry.Wrap(err)

	domain = conf.Domain
	mastodonToken = conf.MastodonToken
	telegraphToken = conf.TelegraphToken
	color.Green("读取配置成功")
}

// 检测是否链接中包含一些敏感参数
// 比如微信公众号的链接包含了 sharer_shareid 参数
func leakSecretInfo(link string) (bool, string) {
	if url, err := url.Parse(link); err == nil {
		host := url.Host
		q := url.Query()
		if host == "mp.weixin.qq.com" {
			return q.Get("sharer_shareid") != "", "警告⚠️你的上条嘟文的微信文章链接中包含 ‘sharer_shareid’ 这样的隐私参数，利用它可以追查到文章的第一位分享者。建议分享“永久短链接”（公众号文章右上角三点-复制链接）。如果您不在乎，请忽略本消息。"
		} else if match, err := regexp.MatchString(".*music.163.com$", host); err == nil && match {
			return q.Get("userid") != "", "警告⚠️你的上条嘟文的网易云音乐链接中包含 ‘userid’参数，利用它可以反查到第一位分享者。建议手动删除该参数。如果您不在乎，请忽略本消息。"
		}
	}
	return false, ""
}
