package police

type NewsResponse struct {
	NewsList          []NewsItem `json:"NewsList"`
	TotalNumberOfNews int        `json:"TotalNumberOfNews"`
}

type NewsItem struct {
	DistrictName       string `json:"DistrictName"`
	ArticleType        string `json:"ArticleType"`
	PublishDate        string `json:"PublishDate"`
	Link               string `json:"Link"`
	ListDate           string `json:"ListDate"`
	Headline           string `json:"Headline"`
	Manchet            string `json:"Manchet"`
	Article            string `json:"Article"`
	ID                 string `json:"Id"`
	ToolTip            string `json:"ToolTip"`
	Image              string `json:"Image"`
	ImageDescription   string `json:"ImageDescription"`
	PhotographerText   string `json:"PhotographerText"`
	NoPhoto            bool   `json:"NoPhoto"`
	DisplayDescription bool   `json:"DisplayDescription"`
}
