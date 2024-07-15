package task

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/VividCortex/ewma"
	"github.com/thehxdev/cfst/utils"
)

const (
	bufferSize                     = 1024
	defaultHttpingURL              = `https://speed.cloudflare.com/`
	defaultURL                     = `https://speed.cloudflare.com/__down?bytes=500000000`
	defaultTimeout                 = 10 * time.Second
	defaultDisableDownload         = false
	defaultTestNum                 = 10
	defaultMinSpeed        float64 = 0.0
)

var (
	URL        = defaultURL
	HttpingURL = defaultHttpingURL
	Timeout    = defaultTimeout
	Disable    = defaultDisableDownload

	TestCount    = defaultTestNum
	MinSpeed     = defaultMinSpeed
	WriteIPsFile = ""
)

func checkDownloadDefault() {
	if URL == "" {
		URL = defaultURL
	}
	if Timeout <= 0 {
		Timeout = defaultTimeout
	}
	if TestCount <= 0 {
		TestCount = defaultTestNum
	}
	if MinSpeed <= 0.0 {
		MinSpeed = defaultMinSpeed
	}
}

func TestDownloadSpeed(ipSet utils.PingDelaySet) (speedSet utils.DownloadSpeedSet) {
	checkDownloadDefault()
	if Disable {
		return utils.DownloadSpeedSet(ipSet)
	}
	if len(ipSet) <= 0 {
		fmt.Println("\n[INFO] The number of IPs in the latency test result is 0, skipping the download speed test...")
		return
	}
	testNum := TestCount
	if len(ipSet) < TestCount || MinSpeed > 0 {
		testNum = len(ipSet)
	}
	if testNum < TestCount {
		TestCount = testNum
	}

	fmt.Printf("Start download speed test (lower limit: %.2f MB/s, number: %d, queue: %d)\n", MinSpeed, TestCount, testNum)

	bar_a := len(strconv.Itoa(len(ipSet)))
	bar_b := "     "
	for i := 0; i < bar_a; i++ {
		bar_b += " "
	}

	bar := utils.NewBar(TestCount, bar_b, "")
	for i := 0; i < testNum; i++ {
		speed := downloadHandler(ipSet[i].IP)
		ipSet[i].DownloadSpeed = speed
		if speed >= MinSpeed*1024*1024 {
			bar.Grow(1, "")
			speedSet = append(speedSet, ipSet[i])
			if len(speedSet) == TestCount {
				break
			}
		}
	}

	bar.Done()
	if len(speedSet) == 0 {
		speedSet = utils.DownloadSpeedSet(ipSet)
	}

	sort.Sort(speedSet)
	return
}

func getDialContext(ip *net.IPAddr) func(ctx context.Context, network, address string) (net.Conn, error) {
	var fakeSourceAddr string
	if isIPv4(ip.String()) {
		fakeSourceAddr = fmt.Sprintf("%s:%d", ip.String(), TCPPort)
	} else {
		fakeSourceAddr = fmt.Sprintf("[%s]:%d", ip.String(), TCPPort)
	}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, fakeSourceAddr)
	}
}

// return download Speed
func downloadHandler(ip *net.IPAddr) float64 {
	client := &http.Client{
		Transport: &http.Transport{DialContext: getDialContext(ip)},
		Timeout:   Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) > 10 {
				return http.ErrUseLastResponse
			}
			if req.Header.Get("Referer") == defaultURL {
				req.Header.Del("Referer")
			}
			return nil
		},
	}
	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return 0.0
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.80 Safari/537.36")

	response, err := client.Do(req)
	if err != nil {
		return 0.0
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return 0.0
	}
	timeStart := time.Now()
	timeEnd := timeStart.Add(Timeout)

	contentLength := response.ContentLength
	buffer := make([]byte, bufferSize)

	var (
		contentRead     int64 = 0
		timeSlice             = Timeout / 100
		timeCounter           = 1
		lastContentRead int64 = 0
	)

	var nextTime = timeStart.Add(timeSlice * time.Duration(timeCounter))
	e := ewma.NewMovingAverage()

	for contentLength != contentRead {
		currentTime := time.Now()
		if currentTime.After(nextTime) {
			timeCounter++
			nextTime = timeStart.Add(timeSlice * time.Duration(timeCounter))
			e.Add(float64(contentRead - lastContentRead))
			lastContentRead = contentRead
		}

		if currentTime.After(timeEnd) {
			break
		}

		bufferRead, err := response.Body.Read(buffer)
		if err != nil {
			if err != io.EOF {
				break
			} else if contentLength == -1 {
				break
			}

			last_time_slice := timeStart.Add(timeSlice * time.Duration(timeCounter-1))
			e.Add(float64(contentRead-lastContentRead) / (float64(currentTime.Sub(last_time_slice)) / float64(timeSlice)))
		}
		contentRead += int64(bufferRead)
	}
	return e.Value() / (Timeout.Seconds() / 120)
}
