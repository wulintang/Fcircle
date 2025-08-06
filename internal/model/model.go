package model

// Friend 代表一个友链的基本信息
type Friend struct {
	Name   string `json:"name"`   // 昵称
	URL    string `json:"url"`    // 个人站地址
	Avatar string `json:"avatar"` // 头像地址
	RSS    string `json:"RSS"`    // RSS 订阅源
}

// Article 代表从 RSS 中抓取的一篇文章
type Article struct {
	Title     string `json:"title"`     // 标题
	Link      string `json:"link"`      // 文章链接
	Published string `json:"published"` // 发布时间
	Author    string `json:"author"`    // 作者昵称
	Avatar    string `json:"avatar"`    // 作者头像
	Content   string `json:"content"`   // 内容
	Url       string `json:"url"`       // 个人站地址
}

// FeedResult 用于输出最终 JSON 文件结构（原有定义）
type FeedResult struct {
	Meta struct {
		FetchTime    string `json:"fetch_time"`    // 抓取时间
		FriendCount  int    `json:"friend_count"`  // 配置中友链数量
		SuccessCount int    `json:"success_count"` // 抓取成功的 RSS 数
		FailCount    int    `json:"fail_count"`    // 抓取失败的 RSS 数
		ArticleCount int    `json:"article_count"` // 总共抓取到的文章数
	} `json:"meta"`
	Items []Article `json:"items"` // 所有抓取到的文章列表
}

// DomainFeedItem 用于按个人站地址分组的文章信息
type DomainFeedItem struct {
	Title  string `json:"title"`  // 文章标题
	Link   string `json:"link"`   // 文章链接
	Source string `json:"source"` // 来源（个人站地址）
	Date   string `json:"date"`   // 发布日期
}

// DomainFeedResult 按个人站地址分组的最终输出格式
type DomainFeedResult map[string][]DomainFeedItem

// ConvertToDomainFormat 将原有FeedResult转换为按个人站地址分组的格式
func (fr *FeedResult) ConvertToDomainFormat() DomainFeedResult {
	result := make(DomainFeedResult)
	
	// 遍历所有文章，按个人站地址分组
	for _, article := range fr.Items {
		// 直接使用个人站地址作为分组键和source值
		siteUrl := article.Url
		if siteUrl == "" {
			continue
		}
		
		// 转换为目标格式的文章结构
		item := DomainFeedItem{
			Title:  article.Title,
			Link:   article.Link,
			Source: siteUrl,  // source直接使用个人站地址
			Date:   article.Published,
		}
		
		// 添加到对应个人站地址的数组中
		result[siteUrl] = append(result[siteUrl], item)
	}
	
	return result
}
    
