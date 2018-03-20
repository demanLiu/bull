package main

import (
	"bytes"
	"flag"
	"fmt"
	"math"
	"net/http"
	"runtime"
	"strings"
	"time"
)

func main() {

	defaultNum := runtime.NumCPU()
	cNum := flag.Int("c", 1, "nums of client")
	requestNum := flag.Int("n", 1, "nums of request")
	cpuNum := flag.Int("cpu", defaultNum, "nums of cpu")
	flag.Parse()
	finallyCpuNum := math.Max(float64(*cpuNum), float64(defaultNum))
	runtime.GOMAXPROCS(int(finallyCpuNum))
	url := flag.Args()[0]
	fmt.Println(*cNum)
	fmt.Println(*requestNum)
	fmt.Println(url)
	for *cNum > 0 {
		go httpget(url, *cNum)
		*cNum--
	}

	time.Sleep(3 * time.Second)
}
func httpget(url string, clientNums int) bool {
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
	fmt.Println(res.StatusCode)
	return res.StatusCode == 200
}
