package main

import (
	"Fcircle/internal/config"
	"Fcircle/internal/fetcher"
	"Fcircle/internal/utils"
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"net/http"
	"os"
	"sync"
)

var (
	fetchMutex sync.Mutex
	isFetching bool
	appConfig  *config.AppConfig
)

func main() {

	_ = os.Setenv("TZ", "Asia/Shanghai")

	appConfig = config.LoadConfig()

	err := utils.InitLog(appConfig.Log.File)
	if err != nil {
		log.Fatalf("日志初始化失败：%v", err)
	}

	fmt.Println("程序启动，开始首次抓取...")
	go fetchAndSave()

	c := cron.New(cron.WithSeconds()) // 支持秒字段的 Cron 表达式
	_, err = c.AddFunc(appConfig.Task.CronExpr, fetchAndSave)
	if err != nil {
		fmt.Println("定时任务添加失败：", err)
		return
	}
	c.Start()

	fmt.Printf("定时任务添加成功，Cron 表达式为： %v\n", appConfig.Task.CronExpr)

	http.HandleFunc("/fetch", httpFetchHandler)

	http.HandleFunc("/feed", httpFeedResultHandler)

	addr := fmt.Sprintf(":%d", appConfig.Server.Port)
	fmt.Printf("HTTP服务启动，监听端口 %d\n", appConfig.Server.Port)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("HTTP服务启动失败: %v\n", err)
	}
}

func fetchAndSave() {
	fetchMutex.Lock()
	if isFetching {
		fmt.Println("抓取任务正在执行中，请稍后...")
		fetchMutex.Unlock()
		return
	}
	isFetching = true
	fetchMutex.Unlock()

	defer func() {
		fetchMutex.Lock()
		isFetching = false
		fetchMutex.Unlock()
	}()

	friends, err := fetcher.LoadRemoteFriends(appConfig.RSS.ConfigURL)
	if err != nil {
		fmt.Printf("加载友链配置失败: %v\n", err)
		os.Exit(1)
	}

	result := fetcher.CrawlArticles(friends)

	err = utils.WriteToFile(appConfig.RSS.OutputFile, result)
	if err != nil {
		fmt.Printf("写入结果文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("抓取完成，共 %d 篇文章，结果写入 %s\n", result.Meta.ArticleCount, appConfig.RSS.OutputFile)
}

func httpFetchHandler(w http.ResponseWriter, r *http.Request) {
	fetchMutex.Lock()
	if isFetching {
		fetchMutex.Unlock()
		w.WriteHeader(http.StatusTooEarly) // 425
		w.Write([]byte("抓取任务正在执行中，请稍后再试"))
		return
	}
	isFetching = true
	fetchMutex.Unlock()

	go func() {
		defer func() {
			fetchMutex.Lock()
			isFetching = false
			fetchMutex.Unlock()
		}()

		fmt.Println("HTTP接口触发抓取任务开始...")
		friends, err := fetcher.LoadRemoteFriends(appConfig.RSS.ConfigURL)
		if err != nil {
			fmt.Println("加载友链配置失败:", err)
			return
		}
		result := fetcher.CrawlArticles(friends)
		err = utils.WriteToFile(appConfig.RSS.OutputFile, result)
		if err != nil {
			fmt.Println("写入结果文件失败:", err)
			return
		}
		fmt.Printf("HTTP触发抓取完成，共 %d 篇文章\n", result.Meta.ArticleCount)
	}()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("抓取任务已启动"))
}

// 新增的接口处理函数，直接返回 feed_result.json 内容
func httpFeedResultHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	filePath := appConfig.RSS.OutputFile
	data, err := os.ReadFile(filePath)
	if err != nil {
		http.Error(w, "读取数据失败，请稍后重试", http.StatusInternalServerError)
		utils.Errorf("读取feed结果文件失败: %v", err)
		return
	}

	w.Write(data)
}
