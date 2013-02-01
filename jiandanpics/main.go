package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"time"
)

const (
	NCPU           = 4
	JSONPATH       = "./info.json"
	constSiteUrl   = "http://jandan.net/pic/page-%d#comments"
	constRe        = `<p><img src="(.*jpg)" />` // how PNG?
	constBSizeId   = 1
	constBSizePage = 1
	constBSizePic  = 5
)

type Item struct {
	Hash []byte
	Url  string
	Size int
}

var info map[string]Item

func main() {
	start := time.Now()

	runtime.GOMAXPROCS(NCPU)
	log.SetFlags(log.Lmicroseconds)
	log.Println("START")

	conf := config{JSONPATH}
	err := conf.Load(&info)
	if err != nil {
		log.Fatal("Load:", err)
		os.Exit(1)
	}

	ids := genPageRange(60, 70)

	pages := genPageUrls(ids)

	pics := parsePicUrls(pages)

	var wgQuit sync.WaitGroup
	//handlePics(pics, wgQuit)
	//logPicsUrl(pics, wgQuit)
	fetchPics3(pics, &wgQuit) //must use pointer 
	wgQuit.Wait()

	err = conf.Save(&info)
	if err != nil {
		log.Fatal("Save:", err)
	}

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

func logPicsUrl(in <-chan string, wg sync.WaitGroup) {
	site := make(map[string]uint64)
	for {
		picUrl := <-in
		if len(picUrl) == 0 { //quit msg
			break
		}
		u, _ := url.Parse(picUrl)
		site[u.Host]++
		log.Println("logPicsUrl:", u.Host, site[u.Host])
	}
	log.Println("Q:logPicsUrl:", site)
}

func fetchPics2(in <-chan string, wg *sync.WaitGroup) {
	var i = 0
	for {
		picUrl := <-in
		if len(picUrl) == 0 { //quit msg
			break
		}
		_, ok := info[picUrl]
		if ok {
			log.Println("PicUrlsExit:", picUrl)
			continue
		}
		log.Println("PicUrls  IN:", picUrl)

		resp, err := http.Get(picUrl)
		if err != nil {
			log.Println("download", err)
			continue
		}
		log.Println("Get Finshed:", picUrl)
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("ioutil.ReadAll:", err)
			continue
		}

		log.Println("Hash       :", picUrl)
		it := Item{Url: picUrl}
		filename := hashPic(body, &it)

		path, _ := filepath.Abs(fmt.Sprintf("./pics/%s.jpg", filename))
		log.Println("Write      :", picUrl, path)
		err = ioutil.WriteFile(path, body, 0)
		if err != nil {
			log.Fatal("ioutil.WriteFile", err)
			continue
		}
		log.Println("Write    OK:", picUrl, path)

		i++
	}
	log.Println("Q:fetchPics")

}

func fetchPics3(in <-chan string, wg *sync.WaitGroup) {
	var i = 0
	for {
		picUrl := <-in
		if len(picUrl) == 0 { //quit msg
			break
		}
		_, ok := info[picUrl]
		if ok {
			log.Println("PicUrlsExit:", picUrl)
			continue
		}
		log.Println("PicUrls  IN:", picUrl)

		func() {
			resp, err := http.Get(picUrl)
			if err != nil {
				log.Println("download", err)
				return
			}
			log.Println("Get Finshed:", picUrl)
			defer resp.Body.Close()

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("ioutil.ReadAll:", err)
				return
			}

			log.Println("Hash       :", picUrl)
			it := Item{Url: picUrl}
			filename := hashPic(body, &it)

			path, _ := filepath.Abs(fmt.Sprintf("./pics/%s.jpg", filename))
			log.Println("Write      :", picUrl, path)
			err = ioutil.WriteFile(path, body, 0)
			if err != nil {
				log.Fatal("ioutil.WriteFile", err)
				return
			}
			log.Println("Write    OK:", picUrl, path)
		}()
		i++
	}
	log.Println("Q:fetchPics")

}

func fetchPics(in <-chan string, wg *sync.WaitGroup) {
	var i = 0
	for {
		picUrl := <-in
		if len(picUrl) == 0 { //quit msg
			break
		}
		_, ok := info[picUrl]
		if ok {
			log.Println("PicUrlsExit:", picUrl)
			continue
		}
		log.Println("PicUrls  IN:", picUrl)

		resp, err := http.Get(picUrl)
		if err != nil {
			log.Println("download", err)
			continue
		}
		log.Println("Get Finshed:", picUrl)

		wg.Add(1)
		go writePic(resp, picUrl, wg)

		i++
	}
	log.Println("Q:fetchPics")
}

func writePic(resp *http.Response, picUrl string, wg *sync.WaitGroup) {
	defer resp.Body.Close()
	defer wg.Done()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ioutil.ReadAll:", err)
		return
	}

	log.Println("Hash       :", picUrl)
	i := Item{Url: picUrl}
	filename := hashPic(body, &i)

	path, _ := filepath.Abs(fmt.Sprintf("./pics/%s.jpg", filename))
	log.Println("Write      :", picUrl, path)
	err = ioutil.WriteFile(path, body, 0)
	if err != nil {
		log.Fatal("ioutil.WriteFile", err)
		return
	}
	log.Println("Write    OK:", picUrl, path)
}

func writeDuck(resp *http.Response, picUrl string, wg *sync.WaitGroup) {
	defer resp.Body.Close()
	defer wg.Done()
}

func hashPic(b []byte, i *Item) string {
	h := md5.New()
	i.Size, _ = io.WriteString(h, string(b))
	i.Hash = h.Sum(nil)
	info[i.Url] = *i
	return fmt.Sprintf("%x", i.Hash)
}
