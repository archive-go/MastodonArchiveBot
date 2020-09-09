package main

type (
	config struct {
		Cookie               string
		Domain string
		MastodonToken        string
		TelegraphToken       string
	}
	// WebSocket的信息流数据类型
	stream struct {
		Event   string      `json:"event"`
		Payload interface{} `json:"payload"`
	}
)
