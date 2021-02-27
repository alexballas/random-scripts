package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

func main() {

	mu := &sync.Mutex{}
	var visited []string
	c := colly.NewCollector(
		colly.AllowedDomains(""),
		colly.MaxDepth(5),
		colly.Async(),
	)
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 1})
	// Find and visit all links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		urltovisit, err := url.Parse(e.Attr("href"))
		if err != nil {
			fmt.Println(err)
		}

		elemtovisit := strings.Split(urltovisit.Path, "/")
		if len(elemtovisit) < 3 {
			elemtovisit = append(elemtovisit, "dummy", "dummy")
		}
		if elemtovisit[1] != "en" {
			if len(elemtovisit) == 5 {
				if elemtovisit[4] == "print.pdf" {
					return
				}
				if elemtovisit[2] == "categories" && elemtovisit[3] != "p" {
					folder := "./sydages/" + elemtovisit[3]
					os.MkdirAll(folder, os.ModePerm)
					mainURL := e.Request.URL.Scheme + "://" + e.Request.URL.Hostname()
					pdfPath := "/el/recipe/" + elemtovisit[4] + "/print.pdf"
					if !contains(visited, pdfPath) {
						mu.Lock()
						visited = append(visited, pdfPath)
						mu.Unlock()
						client := http.Client{}
						resp, err := client.Get(mainURL + pdfPath)
						if err != nil {
							fmt.Println(err)
						}
						var pdfgname string
						if len(resp.Header["Content-Disposition"]) == 1 {
							splitHeader := strings.Split(resp.Header["Content-Disposition"][0], "; ")
							for _, q := range splitHeader {
								q = strings.ReplaceAll(q, "filename=", "")
								if q != "attachment" {
									pdfgname = q[1 : len(q)-1]
								}
							}
						}
						pdfBytes, err := io.ReadAll(resp.Body)
						if err != nil {
							fmt.Println(err)
						}
						mu.Lock()
						pdfgname = strings.ReplaceAll(pdfgname, "/", "-")
						os.WriteFile(folder+"/"+pdfgname, pdfBytes, os.ModePerm)
						mu.Unlock()
						fmt.Println(folder+"/"+pdfgname, elemtovisit[4])
						resp.Body.Close()
					} else {
						fmt.Println("Visited", pdfPath)
					}
					//e.Request.Visit("el/recipe/" + elemtovisit[3] + "/print.pdf")
				}
			}
			e.Request.Visit(e.Attr("href"))
		}
	})

	c.Visit("")
	c.Wait()
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
