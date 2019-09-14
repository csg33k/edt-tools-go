package main

import (
	"edt-tools-go/support"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	dryRun := flag.Bool("dry", false, "dryrun")
	slow := flag.Bool("slow", false, "slow")
	custom := flag.String("custom", "", "a custom config DATA_SOURCE,URL Pattern")
	customFile := flag.String("file", "", "input file")

	flag.Parse()
	params := support.Parameters{DryRun: *dryRun, Slow: *slow, Custom: *custom, InputFile: *customFile}
	if *custom != "" {
		split := strings.Split(*custom, ",")
		integration, valid:= support.IntegrationMappings[split[0]]
		if valid {
			support.IntegrationMappings[support.CUSTOM] = support.Integration{DataSource: split[0], Url: integration.Url}
		} else {
			msg := *custom
			msg = fmt.Sprintf(msg, "data source: %s is not valid", msg)
			panic(msg)
		}
	}

	processUrls(params)
}

func processUrls(params support.Parameters) {
	start := time.Now()
	var dataSource string
	if params.Custom != "" {
		dataSource = support.CUSTOM
	} else {
		dataSource = support.S3
	}

	ids := readFile(params.InputFile)
	urls := make([]string, len(ids))
	ch := make(chan string)
	for i := 0; i < len(ids); i++ {

		var url = buildUrl(dataSource, ids[i])
		if len(url) > 0 {
			urls[i] = url
		}
	}

	i := 0
	for _, url := range urls {
		if params.Slow {
			callUrlSequential(url, i)
		} else {
			go callUrlConcurrent(url, ch, i, params.DryRun) // start a goroutine
		}
		i += 1
	}

	if params.DryRun {
		logUrlsToFile("/tmp/urls.txt", urls, ch)
	} else {
		if !params.Slow {
			for range urls {
				res := <-ch
				fmt.Println(res) // receive from hannel ch

			}
		}
	}

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

func logUrlsToFile(fileName string, urls []string, ch chan string) {
	f, err := os.Create(fileName)
	defer f.Close()
	support.Check(err)
	for range urls {
		var msg = <-ch
		fmt.Println(msg) // receive from channel ch
		f.Write([]byte(msg + "\n"))
	}
}

func buildUrl(integration string, id string) string {
	if len(id) == 0 {
		return ""
	}
	var host = ""

	host = support.IntegrationMappings[integration].Url
	if host == "" {
		host = support.IntegrationMappings[support.S3].Url
	}

	var result = fmt.Sprintf(host, id)
	return result

}

func readFile(in string) []string {
	if in == "" {
		in = support.DefaultInputFile
	}

	dat, err := ioutil.ReadFile(in)
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
