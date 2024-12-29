package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/slimcdk/go-police-news"
)

func prettyPrint(emp ...interface{}) {
	empJSON, err := json.MarshalIndent(emp, "", "  ")
	if err != nil {
		log.Fatalf(err.Error())
	}
	fmt.Println(string(empJSON))
}

func main() {

	p := police.New()
	news, err := p.GetDanishNewsResults(time.Now().AddDate(0, -7, 0), time.Now(), police.AllDistricts()...)
	if err != nil {
		log.Fatal(err)
	}

	for i := range news {
		article, err := p.ParseNewsPage(news[i].Link)
		if err != nil {
			log.Println(err)
			continue
		}

		fmt.Println()
		fmt.Println()
		fmt.Println(news[i].Link)
		fmt.Println(article.Header)
		for _, new := range article.Articles {
			fmt.Println()
			fmt.Println(new.Title)
			if len(new.Description) > 99 {
				fmt.Println(new.Description[:50], "...", new.Description[len(new.Description)-50:])
			} else {
				fmt.Println(new.Description)
			}
		}
	}
}
