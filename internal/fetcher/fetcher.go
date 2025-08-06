package fetcher

import (
	"Fcircle/internal/model"
	"Fcircle/internal/utils"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"sync"
	"time"
)

// LoadRemoteFriends 读取远程 JSON 配置
func LoadRemoteFriends(url string) ([]model.Friend, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("获取远程配置文件失败")
	}

	var friends []model.Friend
	if err := json.NewDecoder(resp.Body).Decode(&friends); err != nil {
		return nil, err
	}
	return friends, nil
}

// CrawlArticles 并发抓取所有友链的前N篇文章，返回FeedResult
func CrawlArticles(friends []model.Friend) model.FeedResult {
	const maxConcurrency = 10
	const maxArticlesPerFriend = 10

	var (
		wg           sync.WaitGroup
		mu           sync.Mutex
		successCount int
		failCount    int
		allArticles  []model.Article
	)

	sem := make(chan struct{}, maxConcurrency)

	for _, friend := range friends {
		wg.Add(1)
		sem <- struct{}{}

		go func(f model.Friend) {
			defer wg.Done()
			defer func() { <-sem }()

			articles, err := FetchFriendArticles(f, maxArticlesPerFriend)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				failCount++
				utils.Infof("抓取 [%s] 失败: %v\n", f.Name, err)
			} else {
				successCount++
				allArticles = append(allArticles, articles...)
				utils.Infof("抓取 [%s] 成功，文章数: %d\n", f.Name, len(articles))
			}
		}(friend)
	}

	wg.Wait()

	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("CST", 8*60*60)
	}

	layout := "2006-01-02 15:04:05"

	sort.Slice(allArticles, func(i, j int) bool {
		t1, err1 := time.ParseInLocation(layout, allArticles[i].Published, loc)
		t2, err2 := time.ParseInLocation(layout, allArticles[j].Published, loc)

		if err1 != nil || err2 != nil {
			return false
		}
		return t1.After(t2)
	})

	var result model.FeedResult
	result.Meta.FetchTime = utils.GetNowTime()
	result.Meta.FriendCount = len(friends)
	result.Meta.SuccessCount = successCount
	result.Meta.FailCount = failCount
	result.Meta.ArticleCount = len(allArticles)
	result.Items = allArticles

	return result
}
