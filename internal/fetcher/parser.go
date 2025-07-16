package fetcher

import (
	"Fcircle/internal/model"
	"Fcircle/internal/utils"
	"fmt"
	"github.com/mmcdole/gofeed"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// FetchFriendArticles 请求并解析指定 friend 的 RSS，返回最新 maxCount 篇文章
func FetchFriendArticles(friend model.Friend, maxCount int) ([]model.Article, error) {

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			ResponseHeaderTimeout: 5 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
		},
	}

	const maxRetry = 2

	var (
		resp *http.Response
		err  error
	)

	start := time.Now()

	for attempt := 0; attempt <= maxRetry; attempt++ {
		resp, err = client.Get(friend.RSS)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	duration := time.Since(start)

	utils.Infof("抓取 [%s] RSS 用时: %v", friend.Name, duration)

	if err != nil {
		return nil, fmt.Errorf("RSS 请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS 请求失败，状态码: %d", resp.StatusCode)
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("RSS 解析失败: %v", err)
	}

	articles := make([]model.Article, 0, maxCount)
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}

	for i, item := range feed.Items {
		if i >= maxCount {
			break
		}

		pubTime := time.Now().In(loc)
		if item.PublishedParsed != nil {
			pubTime = item.PublishedParsed.In(loc)
		} else if item.UpdatedParsed != nil {
			pubTime = item.UpdatedParsed.In(loc)
		}

		formattedTime := pubTime.Format("2006-01-02 15:04:05")

		author := friend.Name
		if item.Author != nil && item.Author.Name != "" {
			author = item.Author.Name
		}

		content := ""
		if item.Content != "" {
			content = item.Content
		} else if item.Description != "" {
			content = item.Description
		}

		// 去除 HTML 标签再截取 200 字符
		plainContent := stripHTMLTags(content)
		shortContent := truncate(strings.TrimSpace(plainContent), 200)

		article := model.Article{
			Title:     item.Title,
			Link:      item.Link,
			Published: formattedTime,
			Author:    author,
			Avatar:    friend.Avatar,
			Content:   shortContent,
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// 去除 HTML 标签的函数
func stripHTMLTags(input string) string {
	re := regexp.MustCompile("<[^>]*>")
	return re.ReplaceAllString(input, "")
}

// 截取指定长度字符（中文不会乱码）
func truncate(str string, length int) string {
	runeStr := []rune(str)
	if len(runeStr) > length {
		return string(runeStr[:length]) + "…"
	}
	return str
}
