package service

import (
	"demo-grpc/server/model"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func CrawlWeb(url string) (*model.OpenGraphModel, error) {
	// fmt.Println("---------------- Start crawl website--------------------")
	// reader := bufio.NewReader(os.Stdin)
	// fmt.Print("Vui lòng nhập URL: ")
	// url, err := reader.ReadString('\n')
	// if err != nil {
	// 	log.Fatal(err)
	// }
	url = strings.TrimSpace(url)
	// Crawl website using http and goquery
	res, err := http.Get(url)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Println("status code error: %d %s", res.StatusCode, res.Status)
		return nil, err
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	openGraphModel := ParseDoc(doc)

	openGraphModel.Filename, openGraphModel.Etag, err = UploadFileToBucket(openGraphModel.Image, res.Header.Get("Content-Type"))
	if err != nil {
		log.Println("get error:", err)
		return nil, err
	}
	return &openGraphModel, nil

	// Write to data to output.json
	// log.Println(string(file))
	// file, _ := json.MarshalIndent(openGraphModel, " ", " ")
	// _ = ioutil.WriteFile("output.json", file, 0644)
}

func ParseDoc(doc *goquery.Document) (openGraphModel model.OpenGraphModel) {
	metaAttr := findMetaAttr(doc)
	doc.Find("meta").Each(func(i int, el *goquery.Selection) {
		// type
		value, _ := el.Attr(metaAttr)
		if strings.Contains(value, "type") {
			openGraphModel.Type, _ = el.Attr("content")
		}
		// Title
		if strings.Contains(value, "title") {
			openGraphModel.Title, _ = el.Attr("content")
		}
		// siteName
		if metaAttr == "name" {
			if strings.Contains(value, "site") {
				openGraphModel.SiteName, _ = el.Attr("content")
			}
		} else if metaAttr == "property" {
			if strings.EqualFold(value, "og:site_name") {
				openGraphModel.SiteName, _ = el.Attr("content")
			}
		}
		// description
		if strings.Contains(value, "description") {
			openGraphModel.Description, _ = el.Attr("content")
		}
		// author
		if strings.Contains(value, "author") {
			openGraphModel.Author, _ = el.Attr("content")
		}
		// image
		if strings.Contains(value, "image") && !strings.Contains(value, "image:") {
			openGraphModel.Image, _ = el.Attr("content")
		}
		// url
		if strings.Contains(value, "url") {
			openGraphModel.Url, _ = el.Attr("content")
		}
	})
	if openGraphModel.Title == "" {
		openGraphModel.Title = doc.Find("title").Text()
	}
	return
}

func findMetaAttr(doc *goquery.Document) (metaAttr string) {
	// property
	doc.Find("meta").Each(func(i int, el *goquery.Selection) {
		value, exists := el.Attr("property")
		if exists {
			if strings.Contains(value, "og:") {
				metaAttr = "property"
				return
			}
		}
	})
	if metaAttr != "" {
		return
	}
	// // name
	// doc.Find("meta").Each(func(i int, el *goquery.Selection) {
	// 	value, exists := el.Attr("name")
	// 	if exists {
	// 		log.Println(i)
	// 		log.Println("name:", value)
	// 		if strings.Contains(value, "og:")  {
	// 			metaAttr = "name"
	// 			return
	// 		}
	// 	}
	// })
	return "name"
}
