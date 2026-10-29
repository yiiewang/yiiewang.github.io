// seed：最小可运行样例 —— 把一个 Markdown 文件里的所有链接打印出来。
//
// 只做一件事，不判断对错。后面的能力都从这个 seed 长出来。
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// 匹配 Markdown 行内链接与图片：
//
//	m[1] == "!" 表示是图片
//	m[2] 是链接文字
//	m[3] 是链接目标
var reLink = regexp.MustCompile(`(!?)\[([^\]]*)\]\(([^)]*)\)`)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go run . <file.md>")
		os.Exit(1)
	}
	path := os.Args[1]

	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	for i, line := range strings.Split(string(data), "\n") {
		for _, m := range reLink.FindAllStringSubmatch(line, -1) {
			fmt.Printf("%s:%d\timage=%v\ttext=%q\ttarget=%q\n",
				path, i+1, m[1] == "!", m[2], m[3])
		}
	}
}
