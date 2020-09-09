package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/fatih/color"
)

// 交叉编译：CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o init *.go
// ./init -cookie="123" -ws-protocol="456" -mastodon-token="789" -telegraph-token="101112"
func createConfig() {
	cookieFlag := flag.String("cookie", "", "长毛象 WebSocket 请求的 Cookie 数据")
	mastodonTokenFlag := flag.String("mastodon-token", "", "mastodon token")
	telegraphTokenFlag := flag.String("telegraph-token", "", "telegraph token")
	flag.Parse()

	color.White("Cookie:%s\nMastodonToken:%s\nTelegraphToken:%s", *cookieFlag, *mastodonTokenFlag, *telegraphTokenFlag)

	file, err := os.OpenFile("./config.json", os.O_RDWR|os.O_CREATE, 0766) //打开或创建文件，设置默认权限
	errHandler("读取配置失败", err)
	defer file.Close()

	configData := config{
		Cookie:         *cookieFlag,
		MastodonToken:  *mastodonTokenFlag,
		TelegraphToken: *telegraphTokenFlag,
	}
	configJSON, err := json.Marshal(configData)
	if err != nil {
		color.Red("序列化config.json失败", err.Error())
		return
	}

	if _, err := file.Write(configJSON); err != nil {
		color.Red("写入数据到 config.json 失败", err.Error())
		return
	}

	color.Green("初始化config.json成功")
}
