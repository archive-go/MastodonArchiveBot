# MastodonArchiveBot
监测长毛象世界的所有嘟文，监测到如果存在某些特定链接（如微信公众号文章）就自动备份然后将备份结果评论到原嘟下


CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mastodon/init init *.go