package main


import (
	"fmt"
	"net/http"
	"strings"
	"net/url"
	"time"
	"math/rand"
	"github.com/PuerkitoBio/goquery"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/61.0.3163.100 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:56.0) Gecko/20100101 Firefox/56.0",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Safari/604.1.38",
}


func randomUserAgent()string{
	rand.Seed(time.Now().Unix())
	rand := rand.Int() % len(userAgents)
	return userAgents[rand]
}

func getResponse(link string, baseUrl string)(*http.Response){
	client := http.Client{}
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", randomUserAgent())

	res, _ := client.Do(req)
	return res
}


func discoverUrls(response *http.Response)[]string{
	if response != nil {
		doc, _ := goquery.NewDocumentFromReader(response.Body)
		foundUrls := []string{}

		if doc != nil {
			doc.Find("a").Each(func(i int, s *goquery.Selection){
				link, _ := s.Attr("href")
				foundUrls = append(foundUrls, link)
			})
		}
		return foundUrls
	} 
	return []string{}
}


func resolveRelative(href string, baseUrl string)string{
	if strings.HasPrefix(href, "/"){
		return fmt.Sprintf("%s%s", baseUrl, href)
	} else{
		return href
	} 
	
}


func resolveRelativeLinks(href string, baseUrl string)(bool, string){
	resultHref := resolveRelative(href, baseUrl)
	urlBase, _ := url.Parse(baseUrl)
	urlHref, _ := url.Parse(resultHref)
	if urlBase != nil && urlHref != nil{
		if urlBase.Host == urlHref.Host {
			return true, resultHref
		}
		return false, ""
	}
	return false, ""
}

var tokens = make(chan struct{}, 5)


func crawl(link string, baseUrl string)[]string{
	fmt.Println(link)
	tokens <- struct{}{}
	res := getResponse(link, baseUrl)
	<- tokens
	hreflinks := discoverUrls(res)
	foundUrls := []string{}

	for _, link := range hreflinks{
		ok, correctLink := resolveRelativeLinks(link, baseUrl)
		if ok{
			if correctLink != ""{
				foundUrls = append(foundUrls, correctLink)
			}
		}
	}

	return foundUrls
}


func main(){
	workList := make(chan []string)
	var n int
	n++
	baseDomain := "https://thegaurdian.com"

	go func(){workList <- []string{baseDomain}}()

	seen := make(map[string]bool)

	for ; n > 0 ; n--{
		links := <- workList
		for _, link := range links{
			if !seen[link] {
				seen[link] = true
				n++
				go func(link string, baseUrl string){
					foundUrls := crawl(link, baseUrl)
					if foundUrls != nil {
						workList <- foundUrls
					}
				}(link, baseDomain)
			}
		}
	}
}