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
	batchSize := flag.Int("batch", 500, "Override batch size")

	flag.Parse()
	params := support.Parameters{DryRun: *dryRun, Slow: *slow, Custom: *custom, InputFile: *customFile, BatchSize: *batchSize}
	if *custom != "" {
		split := strings.Split(*custom, ",")
		integration, valid := support.IntegrationMappings[split[0]]
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
	batchSize := params.BatchSize
	fmt.Printf("Starting Thread URL calls to force tasks for datasource: %s calling %d endpoints" , params.Custom, len(urls))
	for _, url := range urls {
		if params.Slow {
			callUrlSequential(url, i, params.DryRun)
		} else {
			go callUrlConcurrent(url, ch, i, params.DryRun) // start a goroutine
		}
		i += 1
		if i%batchSize == 0 {
			getResponses(params, batchSize, ch)
		}
		if !params.DryRun {
			time.Sleep(100 * time.Millisecond) // throttle
		}

	}

	if i%batchSize != 0 {
		getResponses(params, i%batchSize, ch)
	}

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

func getResponses(params support.Parameters, len int, ch chan string) {
	if !params.Slow {
		for i := 0; i < len; i++ {
			res := <-ch
			fmt.Println(res)

		}
	}

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

func callUrlSequential(url string, i int, dryRun bool) {
	fmt.Printf("%d:  Calling URL: %s", i, url)
	if dryRun {
		fmt.Printf("\n")
		return
	}
	resp, err := http.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch: %v\n", err)
		os.Exit(1)
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	support.Check(err)
	fmt.Printf(" response is: %s\n", b)

}

func callUrlConcurrent(url string, ch chan string, requestCnt int, dryRun bool) {
	if dryRun {
		ch <- fmt.Sprintf("id: %d, called: %s", requestCnt, url)
		return
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
