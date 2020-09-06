package main

import "github.com/MakeGolangGreat/telegraph-go"

const projectLink = "https://github.com/MakeGolangGreat/MastodonArchiveBot"

var attachInfo = &telegraph.NodeElement{
	Tag: "p",
	Children: []telegraph.Node{
		telegraph.NodeElement{
			Tag: "br",
		},
		telegraph.NodeElement{
			Tag: "blockquote",
			Children: []telegraph.Node{
				telegraph.NodeElement{
					Tag:      "strong",
					Children: []telegraph.Node{"本页面由长毛象 "},
				},
				telegraph.NodeElement{
					Tag: "a",
					Attrs: map[string]string{
						"href": "https://alive.bar/web/accounts/50563",
					},
					Children: []telegraph.Node{"@备份Bot"},
				},
				telegraph.NodeElement{
					Tag:      "strong",
					Children: []telegraph.Node{" 备份，"},
				},
				telegraph.NodeElement{
					Tag:      "strong",
					Children: []telegraph.Node{"代码开源："},
				},
				telegraph.NodeElement{
					Tag: "a",
					Attrs: map[string]string{
						"href": "https://github.com/MakeGolangGreat/MastodonArchiveBot",
					},
					Children: []telegraph.Node{"https://github.com/MakeGolangGreat/MastodonArchiveBot"},
				},
			},
		},
	},
}
