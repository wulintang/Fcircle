package fetcher

import (
	"Fcircle/internal/model"
	"fmt"
	"github.com/mmcdole/gofeed"
	"net/http"
	"time"
)

// FetchFriendArticles 请求并解析指定 friend 的 RSS，返回最新 maxCount 篇文章
func FetchFriendArticles(friend model.Friend, maxCount int) ([]model.Article, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(friend.RSS)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("RSS 请求失败，状态码: %d", resp.StatusCode)
	}

	parser := gofeed.NewParser()
	feed, err := parser.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	articles := make([]model.Article, 0, maxCount)

	for i, item := range feed.Items {
		if i >= maxCount {
			break
		}

		pubTime := time.Now()
		if item.PublishedParsed != nil {
			pubTime = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			pubTime = *item.UpdatedParsed
		}

		author := friend.Name
		if item.Author != nil && item.Author.Name != "" {
			author = item.Author.Name
		}

		article := model.Article{
			Title:     item.Title,
			Link:      item.Link,
			Published: pubTime,
			Author:    author,
			Avatar:    friend.Avatar,
		}
		articles = append(articles, article)
	}

	return articles, nil
}
