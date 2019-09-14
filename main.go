package main

import (
	"edt-tools-go/support"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type ChannelData struct {
	msg string
	status string
}

const (
	//s3Host   = "http://edtsitesclr1-5-qa-dca.agkn.net:8080/s3/task/runnow/%s?force=true"
	//hackHost = "http://edtsclr1-1-prod-dca.agkn.net:8080/tradedesk/task/runnow/%s?force=true"
	//hackHost = "http://edttrddsk1-1-prod-dca.agkn.net:8080/tradedesk-v5/task/runnow/%s?force=true"
	hackHost = "http://edtsclr1-1-prod-dca.agkn.net:8080/tradedesk/task/runnow/%s?force=true"
	s3Host     = "http://edtsclr1-1-prod-dca.agkn.net:8080/s3/task/runnow/%s?force=true"
	adFormHost = "http://edtsclr1-1-prod-dca.agkn.net:8080/adform/task/runnow/%s?force=true"
	s3DateHost = "http://edtsitesclr1-1-prod-dca.agkn.net:8080/s3/scheduler/runnow/%s?feedId=569&force=true"
	sftpHost   = "http://edtsitesclr1-5-qa-dca.agkn.net:8080/s3/task/runnow/%s?force=true"
	dcm        = "http://edtsclr1-1-prod-dca.agkn.net:8080/dcm2/task/runnow/%s?force=true"
	tddv5      = "http://edttrddsk1-1-prod-dca.agkn.net:8080/tradedesk-v5/task/runnow/%s?force=true"
)

type Integration int

const (
	DCM Integration = iota
	SFTP
	S3
	S3_DATE
	TDD_V5
	AD_FORM
	HACK
)

func main() {
	dryRun := flag.Bool("dry", false, "dryrun")
	slow := flag.Bool("slow", false, "slow")

	flag.Parse()
	//support.Www()

	processUrls(*dryRun, *slow)
	//processDateUrls(*dryRun, "2019-04-15")
}

func processDateUrls(dryRun bool, date string) {
	start := time.Now()
	fmt.Println("woot")
	// You can get the IDs from the DB, or reading from file but the expectations are that they will be in the same format
	//ids := dbConnect()
	//ids := readFile()
	//var now = time.Now()
	//println(now)

	//urls := make([]string, len(ids))
	//ch := make(chan string)
	//for i := 0; i < len(ids); i++ {
	//	var url = buildUrl(TDD_V5, ids[i])
	//	if len(url) > 0 {
	//		urls[i] = url
	//	}
	//}
	//
	//i := 0
	//for _, url := range urls {
	//	go callUrlConcurrent(url, ch, i, dryRun) // start a goroutine
	//	i += 1
	//}
	//if dryRun {
	//	logUrlsToFile("/tmp/urls.txt", urls, ch)
	//} else {
	//	for range urls {
	//		fmt.Println(<-ch) // receive from channel ch
	//	}
	//}

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

func processUrls(dryRun bool, slow bool) {
	start := time.Now()
	fmt.Println("woot")
	// You can get the IDs from the DB, or reading from file but the expectations are that they will be in the same format
	//ids := dbConnect()
	ids := readFile()
	urls := make([]string, len(ids))
	ch := make(chan string)
	for i := 0; i < len(ids); i++ {
		var url = buildUrl(HACK, ids[i])
		if len(url) > 0 {
			urls[i] = url
		}
	}

	i := 0
	for _, url := range urls {
		if slow {
			callUrlSequential(url, i)
		} else {
			go callUrlConcurrent(url, ch, i, dryRun) // start a goroutine
		}
		i += 1
	}

	if dryRun {
		logUrlsToFile("/tmp/urls.txt", urls, ch)
	} else {
		if !slow {
			for range urls {
				res := <-ch
				fmt.Println(res) // receive from hannel ch

			}
		}
	}

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

func logUrlsToFile(fileName string, urls []string, ch chan string) {
	f, err := os.Create("/tmp/urls.txt")
	defer f.Close()
	support.Check(err)
	for range urls {
		var msg = <-ch
		fmt.Println(msg) // receive from channel ch
		f.Write([]byte(msg + "\n"))
	}
}

func buildUrl(integration Integration, id string) string {
	if len(id) == 0 {
		return ""
	}
	var host = ""
	switch integration {
	case DCM:
		host = dcm
	case SFTP:
		host = sftpHost
	case TDD_V5:
		host = tddv5
	case AD_FORM:
		host = adFormHost
	case S3_DATE:
		host = s3DateHost
	case HACK:
		host = hackHost
	default:
		host = s3Host
	}

	var result = fmt.Sprintf(host, id)
	return result

}

func readFile() []string {

	dat, err := ioutil.ReadFile("./work/config/ids.txt")
	support.Check(err)
	split := strings.Split(string(dat), "\n")
	return split
}

func callUrlSequential(url string, i int) {
	fmt.Printf("%d:  Calling URL: %s", i, url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
		os.Exit(1)
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	support.Check(err)
	fmt.Printf(" response is: %s\n", b)
	//time.Sleep(100 * time.Millisecond) // throttle

}

func callUrlConcurrent(url string, ch chan string, requestCnt int, dryRun bool) {
	if dryRun {
		ch <- fmt.Sprintf("%s", url)
		return
	} else {
		fmt.Printf("Calling url %s concurrently\n", url)
	}
	start := time.Now()
	resp, err := http.Get(url)
	if err != nil {
		ch <- fmt.Sprint(err) // send to channel ch
		return
	}
	nbytes, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close() // don't leak resources
	if err != nil {
		ch <- fmt.Sprintf("while reading %s: %v", url, err)
		return
	}
	//nbytes := "Hello world"
	secs := time.Since(start).Seconds()
	ch <- fmt.Sprintf("id: %d, called: %s executed in: %7.2f and response was: %s", requestCnt, url, secs, nbytes)

}

func call_url(url string, concurrent bool, channels chan string, requestCount int, dryRun bool) {
	if concurrent {
		go callUrlConcurrent(url, channels, requestCount, dryRun)
		//panic("Concurrency not yet implemented")
	} else {
		callUrlSequential(url, 0)
	}

}
