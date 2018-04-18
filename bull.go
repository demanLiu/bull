package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/http/httptrace"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var waitgroup sync.WaitGroup

type Report struct {
	Concurrency       int
	TotalTime         time.Duration
	DNSTime           time.Duration
	BuildConTime      time.Duration
	RequestTime       time.Duration
	WaitRespTime      time.Duration
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
	method := "GET"
	requestTime := *requestNum / *cNum
	waitgroup.Add(*requestNum)
	go spinner(100 * time.Microsecond)
	for requestTime > 0 {
		perTimeNum = *cNum
		for perTimeNum > 0 {
			go func() {
				result <- httpRequest(method, url)
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
	r := &Report{}
	for s := range result {
		if s {
			r.Non2xxNums++
		}
	}

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
func httpRequest(method, url string) bool {
	begin := time.Now()
	if !strings.Contains(url, ":") {
		b := bytes.Buffer{}
		b.WriteString("http://")
		b.WriteString(url)
		url = b.String()
	}
	//http 按照下列顺序进行
	var getConStart, dnsStart, dnsDone, conStart, conDone, getConDone, writeReq, getResp time.Time
	traceCtx := httptrace.WithClientTrace(context.Background(), &httptrace.ClientTrace{
		GetConn: func(hostPort string) {
			// fmt.Printf("Prepare to get a connection for %s.\n", hostPort)
			getConStart = time.Now()
		},
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = time.Now()
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			dnsDone = time.Now()
		},
		ConnectStart: func(network, addr string) {
			// fmt.Printf("Dialing... (%s:%s).\n", network, addr)
			conStart = time.Now()
		},
		ConnectDone: func(network, addr string, err error) {
			if err == nil {
				// fmt.Printf("Dial is done. (%s:%s)\n", network, addr)
				conDone = time.Now()
			} else {
				fmt.Printf("Dial is done with error: %s. (%s:%s)\n", err, network, addr)
			}
		},
		GotConn: func(info httptrace.GotConnInfo) {
			// fmt.Printf("Got a connection: reused: %v, from the idle pool: %v.\n", info.Reused, info.WasIdle)
			getConDone = time.Now()

		},
		WroteRequest: func(info httptrace.WroteRequestInfo) {
			if info.Err == nil {
				// fmt.Println("Wrote a request: ok.")
				writeReq = time.Now()
			} else {
				fmt.Println("Wrote a request:", info.Err.Error())
			}
		},
		GotFirstResponseByte: func() {
			// fmt.Println("Got the first response byte.")
			getResp = time.Now()
		},
	})
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		log.Fatal("Fatal error:", err)
	}
	req = req.WithContext(traceCtx)
	//TODO client add option
	requestHeader := make(http.Header)
	// requestHeader.Set("Content-Type", "text/html")
	// requestHeader.Set("User-Agent", "lh/0.0.1")
	req.Header = requestHeader
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request error: %v\n", err)
		os.Exit(1)
	}
	dnsTime := dnsDone.Sub(dnsStart)
	buildConTime := conDone.Sub(conStart)
	reqTime := writeReq.Sub(getConDone)
	waitRespTime := getResp.Sub(writeReq)
	totalTime := time.Now().Sub(begin)
	fmt.Println(dnsTime)
	fmt.Println(buildConTime)
	fmt.Println(reqTime)
	fmt.Println(waitRespTime)
	fmt.Println(totalTime)
	fmt.Println("resp")
	fmt.Println(resp.StatusCode)
	var size int64
	size = resp.ContentLength
	defer resp.Body.Close()
	if resp.ContentLength < 0 {
		bodyLength, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
		}
		size = int64(len(bodyLength))
	}
	fmt.Println(size)

	return true
}
