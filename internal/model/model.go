package model

// FeedItem 代表一篇文章的信息
type FeedItem struct {
	Title  string `json:"title"`  // 文章标题
	Link   string `json:"link"`   // 文章链接
	Source string `json:"source"` // 来源网站
	Date   string `json:"date"`   // 发布日期
}

// FeedResult 用于输出最终JSON文件结构
// 键为域名（如"20060611.xyz"），值为该域名下的文章数组
type FeedResult map[string][]FeedItem

// 辅助结构体，用于在处理过程中存储元数据
// 最终生成JSON时可能不需要包含这些元数据
type FeedMetadata struct {
	FetchTime    string // 抓取时间
	FriendCount  int    // 配置中友链数量
	SuccessCount int    // 抓取成功的RSS数
	FailCount    int    // 抓取失败的RSS数
	ArticleCount int    // 总共抓取到的文章数
}
