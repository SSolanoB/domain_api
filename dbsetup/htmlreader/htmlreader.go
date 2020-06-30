package htmlreader

import (
	"fmt"
	"github.com/valyala/fasthttp"
	"strings"
	"github.com/PuerkitoBio/goquery"
	"regexp"
)

func RequestHeaderInfo(url string) (titles, images, links []string, isdown bool) {

 	client := fasthttp.Client{}
  var dst []byte
  status, body, err := client.Get(dst, url)

  if err != nil {
    fmt.Printf("Request failed: %s\n", err)
    fmt.Println(status)
    isdown = true
    fmt.Println(isdown)
    return
  }

  body_io := strings.NewReader(string(body))

  doc, erro := goquery.NewDocumentFromReader(body_io)

  fmt.Println(doc)

  if erro != nil {
  	fmt.Println("Error")
  }

  doc.Find("title").Each(func(i int, s *goquery.Selection) {
    titles = append(titles, s.Text())
  })

  fmt.Println(titles)

  doc.Find("img").Each(func(i int, s *goquery.Selection) {
    images = append(images, s.AttrOr("src", "Not defined"))
  })

  fmt.Println(images)

  doc.Find("link").Each(func(i int, s *goquery.Selection) {
    val, _ := s.Attr("rel")

    re := regexp.MustCompile(`([Ii]con)`)
		ic := re.FindAllStringSubmatch(string(val), -1)
		var ic_1 string
		if ic != nil {
			ic_1 = ic[0][0]
		}

		if ic_1 != "" {
			fmt.Println(ic_1)
			links = append(links, s.AttrOr("href", "Not defined"))
		}
  })

  fmt.Println(links)
  return
}