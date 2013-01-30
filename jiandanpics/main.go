package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"
)

const (
	NCPU           = 4
	constSiteUrl   = "http://jandan.net/pic/page-%d#comments"
	constRe        = `<p><img src="(.*jpg)" />`
	constBSizeId   = 1
	constBSizePage = 1
	constBSizePic  = 5
)

var wgQuit sync.WaitGroup

func main() {
	start := time.Now()

	runtime.GOMAXPROCS(NCPU)
	log.SetFlags(log.Lmicroseconds)
	log.Println("START")

	ids := genPageRange(690, 700)

	pages := genPageUrls(ids)

	pics := parsePicUrls(pages)

	fetchPics(pics)
	wgQuit.Wait()

	log.Println("END:", time.Since(start))
}

func genPageRange(low, high int) <-chan int {
	c := make(chan int, constBSizeId)
	go func() {
		for i := low; i < high; i++ {
			c <- i
		}
		c <- -1
		log.Println("Q:genID")
	}()

	return c
}

func genPageUrls(ids <-chan int) <-chan string {
	c := make(chan string, constBSizePage)
	go func() {
		for {
			id := <-ids
			if id == -1 {
				break
			}

			c <- fmt.Sprintf(constSiteUrl, id)
		}
		c <- string("") //quit msg
		log.Println("Q:genPageUrls")
	}()
	return c
}

func parsePicUrls(in <-chan string) <-chan string {
	out := make(chan string, constBSizePic)
	go func() {
		//var wg sync.WaitGroup

		for {
			page := <-in
			if len(page) == 0 { //quit msg
				break
			}
			log.Println("PageUrls  IN:", page)

			func() {
				resp, err := http.Get(page)
				if err != nil {
					log.Println("http.Get:", err)
					return
				}
				defer resp.Body.Close()

				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println("ioutil.ReadAll:", err)
					return
				}

				//search img urls in html body
				re, err := regexp.Compile(constRe)
				if err != nil {
					log.Println("regexp.Compile:", err)
					return
				}

				lis := re.FindAllSubmatch(body, -1)
				for _, li := range lis {
					out <- fmt.Sprint(string(li[1]))
					log.Println("PicUrls OUT")
				}
			}()
		}

		out <- "" //quit msg
		log.Println("Q:parsePicUrls")
	}()

	return out
}

func fetchPics(in <-chan string) {
	var i = 0
	for {
		picUrl := <-in
		if len(picUrl) == 0 { //quit msg
			break
		}
		log.Println("PicUrls  IN:", picUrl)

		resp, err := http.Get(picUrl)
		if err != nil {
			log.Println("download", err)
			continue
		}
		log.Println("Get Finshed:", picUrl)

		wgQuit.Add(1)
		go writePic(resp, picUrl, i)

		i++
	}
	log.Println("Q:fetchPics")
}

func writePic(resp *http.Response, picUrl string, i int) {
	defer resp.Body.Close()
	defer wgQuit.Done()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ioutil.ReadAll:", err)
		return
	}

	path, _ := filepath.Abs(fmt.Sprintf("./pics/%d.jpg", i))
	log.Println("Write      :", picUrl, path)
	err = ioutil.WriteFile(path, body, 0)
	if err != nil {
		log.Println("ioutil.WriteFile", err)
		return
	}
	log.Println("Write    OK:", picUrl, path)
}

func writeDuck(resp *http.Response, picUrl string, i int) {
	defer resp.Body.Close()
	defer wgQuit.Done()
}
