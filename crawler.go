package main

import (
	"bufio"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
    "time"
)

const PREFIX = "http://baike.baidu.com"
const IMG_PATH = "./images/"

func saveImage(name string, url string) {
	fmt.Println("image url:" + url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Get " + url + " Error!")
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	suffixs := strings.Split(url, ".")
	suffix := suffixs[len(suffixs)-1]
	img_file := IMG_PATH + name + "." + suffix
	fmt.Println("image file name:" + img_file)
	ioutil.WriteFile(img_file, body, 0644)
}

func getFirstImage(title string, url string, c chan string) {
	fmt.Println("url:" + url)
	url = strings.Split(strings.Replace(url, "picview", "picture", 1), "?")[0]
	fmt.Println("Move to:" + url)
	doc, e := goquery.NewDocument(url)
        if e != nil{
		    fmt.Println("Get Image URL From " + url + " Error!")
                c <- ""
		    return
        }
	img := doc.Find("#imgPicture")
        fmt.Println(img)

	img_url, err := doc.Find("#imgPicture").Attr("src")
	if !err {
		fmt.Println("Get Image URL From " + url + " Error!")
                c <- ""
		return
	}
	//saveImage(title, img_url)
        time.Sleep(1000 * time.Millisecond)
	c <- img_url
}

func main() {

	os.Mkdir(IMG_PATH, 0777)

	var record []string
	var flag string
	f, err := os.Open(IMG_PATH + "image_url.dat")
	if err != nil {
		fmt.Print("File " + "image_url.dat" + " open Error")
	} else {
		r := bufio.NewReader(f)
		for line, e := r.ReadString('\n'); e != io.EOF; line, e = r.ReadString('\n') {
			record = append(record, line)
		}
		flag = strings.Split(record[len(record)-1], ":")[0]
	}
	fmt.Println("Flag:" + flag)
	fmt.Println("Flag len:" + string(len(flag)))
	fmt.Println("Flag:" + flag)
	f.Close()

	//fr, rerr := os.Open("./test_data/test.dat")
	fr, rerr := os.Open("/home/xlore/NewBaidu/etc/baidu-dump-20140910.dat")
	fw, werr := os.Create(IMG_PATH + "image_url.dat")
	if rerr != nil || werr != nil {
		fmt.Print("File open Error")
		return
	}
	defer fr.Close()
	defer fw.Close()

	reader := bufio.NewReader(fr)
	writer := bufio.NewWriter(fw)

	var title, url, img_url string
	var hasFirst bool

	for _, line := range record {
		fmt.Fprint(writer, line)
	}
	writer.Flush()

	if len(flag) > 0 {
		end := false
		for line, e := reader.ReadString('\n'); e != io.EOF; line, e = reader.ReadString('\n') {

			items := strings.SplitN(strings.Trim(line, "\n"), ":", 2)
			if len(items) < 2 {
				if end {
					break
				} else {
					continue
				}
			}
			prefix := items[0]
			value := items[1]
			//fmt.Println("Prefix:" + prefix)
			if prefix == "Title" {
				title = strings.TrimSpace(value)
				if title == flag {
					end = true
				}
			}
		}
	}

	for line, e := reader.ReadString('\n'); e != io.EOF; line, e = reader.ReadString('\n') {

		items := strings.SplitN(strings.Trim(line, "\n"), ":", 2)
		if len(items) < 2 {
			continue
		}
		prefix := items[0]
		value := items[1]
		//fmt.Println("Prefix:" + prefix)

		switch prefix {

		case "ID":
			fmt.Println("ID:" + value)

		case "Title":
			title = strings.TrimSpace(value)
			hasFirst = false
			fmt.Println("Title:" + value)

		case "FirstImage":
			hasFirst = true
			url = value[3 : len(value)-2]
			var img_url string
			if strings.HasPrefix(url, "http") {
				//go saveImage(title, url)
				img_url = url
			} else {
				url = PREFIX + url
				c := make(chan string)
				go getFirstImage(title, url, c)
                select {
                    case <-c:
				        img_url = <-c
                        fmt.Println("done")
                    case <-time.After(3 * time.Second):
                        fmt.Println("timeout")
                    }
			}
			fmt.Println("URL:" + img_url)
			fmt.Fprintln(writer, title+":"+img_url)
			writer.Flush()

		case "Images":
			if hasFirst {
				break
			}
			hasFirst = false
			urls := strings.Split(value, "::;")
			items := strings.SplitN(urls[0], "||", 2)
			img_url = items[1][:len(items[1])-2]
			//go saveImage(title, img_url)
			fmt.Println("URL:" + img_url)
			fmt.Fprintln(writer, title+":"+img_url)
			writer.Flush()

		}

	}

	fmt.Println("END")
}
