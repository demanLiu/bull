package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

var waitgroup sync.WaitGroup

type Report struct {
	Concurrency       int
	TotalTime         time.Duration
	CompleteNums      int
	FailedNums        int
	Non2xxNums        int
	TotalTransferred  int
	HtmlTransferred   int
	RequestPerSecond  float32
	TimePerRequest    float32
	TimePerRequestAll float32
	TranferRate       float32
}

func main() {
	var perTimeNum int
	result := make(chan bool)
	defaultNum := runtime.NumCPU()
	cNum := flag.Int("c", 1, "nums of client")
	requestNum := flag.Int("n", 1, "nums of request")
	cpuNum := flag.Int("cpu", defaultNum, "nums of cpu")
	flag.Parse()
	finallyCpuNum := math.Max(float64(*cpuNum), float64(defaultNum))
	runtime.GOMAXPROCS(int(finallyCpuNum))
	url := flag.Args()[0]
	requestTime := *requestNum / *cNum
	waitgroup.Add(*requestNum)

	for requestTime > 0 {
		perTimeNum = *cNum
		for perTimeNum > 0 {
			go func() {
				result <- httpGet(url)
				waitgroup.Done()
			}()
			perTimeNum--
		}
		requestTime--
	}

	go func() {
		waitgroup.Wait()
		close(result)
	}()
	//break until close channel

	for s := range result {
		if s {

		}
	}

}
func httpGet(url string) bool {
	if !strings.Contains(url, ":") {
		b := bytes.Buffer{}
		b.WriteString("http://")
		b.WriteString(url)
		url = b.String()
	}
	res, err := http.Get(url)
	if err != nil {
		// handle error
		fmt.Println(err)

	}
	return res.StatusCode == 200
}

//active
func spinner(delay time.Duration) {
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
	}
}
