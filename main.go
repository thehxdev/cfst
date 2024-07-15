package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/thehxdev/cfst/task"
	"github.com/thehxdev/cfst/utils"
)

const version = "0.1.0"

func init() {
	var printVersion bool
	var help = `CFST ` + version + `
Test the delay and speed of all IPs of Cloudflare CDN to get the fastest IP (IPv4+IPv6)!
https://github.com/thehxdev/cfst

Usage:
    -n 200
        Count of IPs that are proccessed concurrently (default 200 at most 1000)
        Note that higher values cause lower accuracy but test will be faster.
        It's recommended to set [-n] value to 10% of your total IPs.

    -t 4
        Number of packets sent to a single IP address for ping test (default 4)

    -dn 10
        Maximum returned IPs after speed test and sorting operation (default 10)

    -dt 10
        The longest time to download from a single IP (default 10 seconds)

    -tp 443
        Target port (default port 443)

    -url https://speed.cloudflare.com/__down?bytes=500000000
        Speed test address (Download speed test)

    -httping
        Change ping test mode to the HTTP protocol (default false)

    -httping-url https://speed.cloudflare.com/
        Httping measurement address

    -httping-code 200
        HTTP status code returned by the webpage (default 200 301 302)

    -tl 200
        The average delay upper limit (default 9999 ms)

    -tll 40
        The average delay lower limit (default 0 ms)

    -tlr 0.2
        The upper limit of the package loss. the range is 0.00 ~ 1.00 (default 1.00)

    -sl 5
        Download speed limit (default 0.00 mb/s)

    -write-ips
        Write the IPs that passed the ping test (default empty)

    -f ./ip.txt
        IP segments file. If the file contains spaces, please add quotes. Support other CDN IP segments. (default ./ip.txt)

    -ip 1.1.1.1,2.2.2.2/24,2606:4700::/32
        Directly specify the IP segments. Seperated by commas (default empty)

    -o result.csv
        Path to output file (default result.csv)

    -dd
        Disable download speed test. IPs will be sorted by ping test (default false)

    -allip
        Check all IP addresses (Only IPv4) in the IP segments. Default behavior is choose a random IP in each /24 range.

    -version
        Print version

    -h
        Show this help message
`
	var minDelay, maxDelay, downloadTime int
	var maxLossRate float64
	flag.IntVar(&task.Routines, "n", 200, "")
	flag.IntVar(&task.PingTimes, "t", 4, "")
	flag.IntVar(&task.TestCount, "dn", 10, "")
	flag.IntVar(&downloadTime, "dt", 10, "")
	flag.IntVar(&task.TCPPort, "tp", 443, "")
	flag.StringVar(&task.URL, "url", `https://speed.cloudflare.com/__down?bytes=500000000`, "")
	flag.StringVar(&task.HttpingURL, "httping-url", `https://speed.cloudflare.com/`, "")

	flag.BoolVar(&task.Httping, "httping", false, "")
	flag.StringVar(&task.WriteIPsFile, "write-ips", "", "")
	flag.IntVar(&task.HttpingStatusCode, "httping-code", 0, "")
	// flag.StringVar(&task.HttpingCFColo, "cfcolo", "", "")

	flag.IntVar(&maxDelay, "tl", 9999, "")
	flag.IntVar(&minDelay, "tll", 0, "")
	flag.Float64Var(&maxLossRate, "tlr", 1, "")
	flag.Float64Var(&task.MinSpeed, "sl", 0, "")

	flag.StringVar(&task.IPFile, "f", "ip.txt", "")
	flag.StringVar(&task.IPText, "ip", "", "")
	flag.StringVar(&utils.Output, "o", "result.csv", "")

	flag.BoolVar(&task.Disable, "dd", false, "")
	flag.BoolVar(&task.TestAll, "allip", false, "")

	flag.BoolVar(&printVersion, "v", false, "")
	flag.Usage = func() { fmt.Print(help) }
	flag.Parse()

	if task.MinSpeed > 0 && time.Duration(maxDelay)*time.Millisecond == utils.InputMaxDelay {
		fmt.Println("[Tip] When using the [-sl] parameter, it is recommended to match the [-tl] parameter to avoid always measure the speed because the number of [-dn] is not available ...")
	}

	utils.InputMaxDelay = time.Duration(maxDelay) * time.Millisecond
	utils.InputMinDelay = time.Duration(minDelay) * time.Millisecond
	utils.InputMaxLossRate = float32(maxLossRate)
	task.Timeout = time.Duration(downloadTime) * time.Second
	// task.HttpingCFColomap = task.MapColoMap()

	if printVersion {
		println(version)
		os.Exit(0)
	}
}

func main() {
	fmt.Printf("# CFST %s \n\n", version)

	pingData := task.NewPing().Run().FilterDelay().FilterLossRate()

	speedData := task.TestDownloadSpeed(pingData)

	utils.ExportCsv(speedData)
	// speedData.Print()

	endPrint()
}

func endPrint() {
	if runtime.GOOS == "windows" {
		fmt.Printf("Press the Enter or Ctrl+C to exit...")
		fmt.Scanln()
	}
}
