package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"
	"time"
)

const (
	constSiteUrl = "http://jandan.net/pic/page-%d#comments"
	constRe      = `<p><img src="(.*jpg)" />`
)

func main() {
	runtime.GOMAXPROCS(4)

	start := time.Now()

	pages := genPageUrls(1, 10)

	pics := parsePicUrlsInPage(pages)

	fetchPics(pics)
	fmt.Println(time.Since(start))
}

func genPageUrls(begin, end int) <-chan string {
	c := make(chan string)
	go func() {
		for p := begin; p < end; p++ {
			c <- fmt.Sprintf(constSiteUrl, p)
		}
		c <- string("")
	}()
	return c
}

func parsePicUrlsInPage(in <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for {
			page := <-in
			if len(page) == 0 {
				break
			}
			fmt.Println(page)

			resp, err := http.Get(page)
			if err != nil {
				fmt.Print("http.Get:", err)
				//TODO out
				continue
			}

			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Print("ioutil.ReadAll:", err)
				continue
			}

			//search img urls in html body
			re, err := regexp.Compile(constRe)
			if err != nil {
				fmt.Print("regexp.Compile:", err)
				continue
			}

			lis := re.FindAllSubmatch(body, -1)
			fmt.Println(len(lis))
			for _, li := range lis {
				out <- fmt.Sprint(string(li[1]))
			}
		}
		out <- ""
	}()

	return out
}

func fetchPics(in <-chan string) {
	var i = 0
	for {
		picUrl := <-in
		if len(picUrl) == 0 {
			break
		}
		resp, err := http.Get(picUrl)
		if err != nil {
			fmt.Println("download", err)
			continue
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Print("ioutil.ReadAll:", err)
			continue
		}

		path, _ := filepath.Abs(fmt.Sprintf("./pics/%d.jpg", i))
		fmt.Println(path, picUrl)
		err = ioutil.WriteFile(path, body, 0)
		if err != nil {
			fmt.Println("ioutil.WriteFile", err)
			continue
		}

		i++
	}
}
