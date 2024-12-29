package police

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/cascadia"
	"github.com/go-resty/resty/v2"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

const basePath string = "politi.dk/api/news/getNewsResults"

type District string

const (
	DistrictBornholm                       District = "Bornholms Politi"
	DistrictFyn                            District = "Fyns Politi"
	DistrictMidtOgVestsjaelland            District = "Midt og Vestsjaellands Politi"
	DistrictNordjylland                    District = "Nordjyllands Politi"
	DistrictNordsjaelland                  District = "Nordsjaellands Politi"
	DistrictSydsjaellandsOgLollandFalsters District = "Sydsjaellands og Lolland-Falsters Politi"
	DistrictOEstjyllands                   District = "OEstjyllands Politi"
)

type Client interface {
	GetDanishNewsResults(from, to time.Time, districts ...District) ([]NewsItem, error)
	GetDanishNewsResultsPage(pageSize, pageIndex int, from, to time.Time, districts ...District) (NewsResponse, error)
}

type NewsResponse struct {
	NewsList          []NewsItem `json:"NewsList"`
	TotalNumberOfNews int        `json:"TotalNumberOfNews"`
}

type NewsItem struct {
	DistrictName       string    `json:"DistrictName"`
	ArticleType        string    `json:"ArticleType"`
	PublishDate        string    `json:"PublishDate"`
	Link               string    `json:"Link"`
	ListDate           time.Time `json:"ListDate"`
	Headline           string    `json:"Headline"`
	Manchet            string    `json:"Manchet"`
	Article            string    `json:"Article"`
	ID                 string    `json:"Id"`
	ToolTip            string    `json:"ToolTip"`
	Image              string    `json:"Image"`
	ImageDescription   string    `json:"ImageDescription"`
	PhotographerText   string    `json:"PhotographerText"`
	NoPhoto            bool      `json:"NoPhoto"`
	DisplayDescription bool      `json:"DisplayDescription"`
}

type Article struct {
	Header   string
	Articles []NewsArticle
}

type NewsArticle struct {
	Title, Description string
}

type client struct {
	resty *resty.Client
	log   *log.Logger
}

// Base struct
func New() client {
	r := resty.New()
	r.SetBaseURL(basePath)
	return client{resty: r, log: nil}
}

func AllDistricts() []District {
	return []District{DistrictBornholm,
		DistrictFyn,
		DistrictMidtOgVestsjaelland,
		DistrictNordjylland,
		DistrictNordsjaelland,
		DistrictSydsjaellandsOgLollandFalsters,
		DistrictOEstjyllands,
	}
}

func (c *client) GetDanishNewsResultsPage(pageSize, pageIndex int, from, to time.Time, districts ...District) (NewsResponse, error) {

	_districts := make([]string, len(districts))
	for i, district := range districts {
		_districts[i] = strings.ReplaceAll(string(district), " ", "-")
	}

	var data NewsResponse
	var queryError struct {
		Message string `json:"Message"`
	}

	res, err := c.resty.R().
		SetQueryParams(map[string]string{
			"districtQuery": strings.Join(_districts, ","),
			"itemId":        "90deb0b1-8df0-4a2d-823b-cfd7a5add85f",
			"newsType":      "D%C3%B8gnrapporter",
			"isNewsList":    strconv.FormatBool(true),
			"fromDate":      from.Format(time.RFC3339),
			"toDate":        to.Format(time.RFC3339),
			"page":          fmt.Sprintf("%d", pageIndex),
			"pageSize":      fmt.Sprintf("%d", pageSize),
		}).
		SetHeader("Accept", "application/json").
		SetResult(&data).
		SetError(&queryError).
		Get("https://politi.dk/api/news/getNewsResults")
	if err != nil {
		return NewsResponse{}, err
	}

	if res.StatusCode() != http.StatusOK {
		return NewsResponse{}, errors.New(res.Status())
	}

	return data, nil
}

func (c *client) GetDanishNewsResults(from, to time.Time, districts ...District) ([]NewsItem, error) {
	data := make([]NewsItem, 0)
	readFullPage := false
	for i := 0; !readFullPage; i++ {
		d, err := c.GetDanishNewsResultsPage(1000000000, i, from, to, districts...)
		if err != nil {
			return nil, err
		}
		data = append(data, d.NewsList...)
		readFullPage = (len(data) >= d.TotalNumberOfNews)
	}
	return data, nil
}

func (c *client) ParseNewsPage(link string) (Article, error) {

	res, err := c.resty.R().Get(link)
	if err != nil {
		return Article{}, err
	}
	if res.StatusCode() != http.StatusOK {
		return Article{}, errors.New(res.Status())
	}

	doc, err := html.Parse(io.NewSectionReader(strings.NewReader(string(res.Body())), 0, int64(len(res.Body()))))
	if err != nil {
		return Article{}, err
	}

	article := cascadia.MustCompile(".newsArticle").MatchFirst(doc)
	head := cascadia.MustCompile(".news-manchet").MatchFirst(article)
	body := cascadia.MustCompile(".rich-text").MatchFirst(article)
	articles := cascadia.MustCompile("h3").MatchAll(body)

	data := Article{
		Header: head.FirstChild.Data,
	}

	for _, article := range articles {
		title := article.FirstChild.Data
		article.RemoveChild(article.FirstChild)

		text := ""

		for article.NextSibling != nil {
			if article.NextSibling.DataAtom == atom.H3 {
				break
			}
			if article.NextSibling.DataAtom == atom.Br {
				text += " "
				article.Parent.RemoveChild(article.NextSibling)
				continue
			}
			text += article.NextSibling.Data
			article.Parent.RemoveChild(article.NextSibling)
		}

		// Text cleanup
		space := regexp.MustCompile(`\s+`)

		title = strings.ReplaceAll(title, "\n", "")
		title = strings.ReplaceAll(title, "\t", "")
		title = strings.TrimSpace(title)
		title = space.ReplaceAllString(title, " ")
		if strings.Compare(string(title[len(title)-1]), ":") == 0 {
			title = title[:len(title)-1]
		}

		text = strings.ReplaceAll(text, "\n", "")
		text = strings.ReplaceAll(text, "\t", "")
		text = strings.TrimSpace(text)
		text = space.ReplaceAllString(text, " ")

		data.Articles = append(data.Articles, NewsArticle{
			Title:       title,
			Description: text, //strings.ReplaceAll(text, "br", "\r"),
		})
	}

	return data, nil
}
