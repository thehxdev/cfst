package task

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/thehxdev/cfst/utils"
)

const (
	tcpConnectTimeout = time.Second * 1
	maxRoutine        = 1000
	defaultRoutines   = 200
	defaultPort       = 443
	defaultPingTimes  = 4
)

var (
	Routines      = defaultRoutines
	TCPPort   int = defaultPort
	PingTimes int = defaultPingTimes
)

type Ping struct {
	wg      *sync.WaitGroup
	m       *sync.Mutex
	ips     []*net.IPAddr
	csv     utils.PingDelaySet
	control chan bool
	bar     *utils.Bar
	buff    *bytes.Buffer
}

func checkPingDefault() {
	if Routines <= 0 {
		Routines = defaultRoutines
	}
	if TCPPort <= 0 || TCPPort >= 65535 {
		TCPPort = defaultPort
	}
	if PingTimes <= 0 {
		PingTimes = defaultPingTimes
	}
}

func NewPing() *Ping {
	checkPingDefault()
	ips := loadIPRanges()
	return &Ping{
		wg:      &sync.WaitGroup{},
		m:       &sync.Mutex{},
		ips:     ips,
		csv:     make(utils.PingDelaySet, 0),
		control: make(chan bool, Routines),
		bar:     utils.NewBar(len(ips), "Available:", ""),
		buff:    &bytes.Buffer{},
	}
}

func (p *Ping) Run() utils.PingDelaySet {
	if len(p.ips) == 0 {
		return p.csv
	}

	if Httping {
		fmt.Printf("Start HTTPING test (Port: %d, Delay range: %v ~ %v ms, Packet loss: %.2f)\n", TCPPort, utils.InputMinDelay.Milliseconds(), utils.InputMaxDelay.Milliseconds(), utils.InputMaxLossRate)
	} else {
		fmt.Printf("Start TCPING test (Port: %d, Delay range: %v ~ %v ms, Packet loss: %.2f)\n", TCPPort, utils.InputMinDelay.Milliseconds(), utils.InputMaxDelay.Milliseconds(), utils.InputMaxLossRate)
	}

	for _, ip := range p.ips {
		p.wg.Add(1)
		p.control <- false
		go p.start(ip)
	}

	p.wg.Wait()
	p.bar.Done()

	if WriteIPsFile != "" {
		err := os.WriteFile(WriteIPsFile, p.buff.Bytes(), os.FileMode(0644))
		if err != nil {
			fmt.Println("Failed to write tested IPs")
		}
	}

	sort.Sort(p.csv)
	return p.csv
}

func (p *Ping) start(ip *net.IPAddr) {
	defer p.wg.Done()
	p.tcpingHandler(ip)
	<-p.control
}

// bool connectionSucceed float32 time
func (p *Ping) tcping(ip *net.IPAddr) (bool, time.Duration) {
	startTime := time.Now()
	var fullAddress string
	if isIPv4(ip.String()) {
		fullAddress = fmt.Sprintf("%s:%d", ip.String(), TCPPort)
	} else {
		fullAddress = fmt.Sprintf("[%s]:%d", ip.String(), TCPPort)
	}
	conn, err := net.DialTimeout("tcp", fullAddress, tcpConnectTimeout)
	if err != nil {
		return false, 0
	}
	defer conn.Close()
	duration := time.Since(startTime)
	return true, duration
}

// pingReceived pingTotalTime
func (p *Ping) checkConnection(ip *net.IPAddr) (recv int, totalDelay time.Duration) {
	if Httping {
		recv, totalDelay = p.httping(ip)
		return
	}
	for i := 0; i < PingTimes; i++ {
		if ok, delay := p.tcping(ip); ok {
			recv++
			totalDelay += delay
		}
	}
	return
}

func (p *Ping) appendIPData(data *utils.PingData) {
	p.m.Lock()
	defer p.m.Unlock()
	p.csv = append(p.csv, utils.CloudflareIPData{
		PingData: data,
	})
}

// handle tcping
func (p *Ping) tcpingHandler(ip *net.IPAddr) {
	recv, totalDlay := p.checkConnection(ip)
	nowAble := len(p.csv)
	if recv != 0 {
		nowAble++
	}
	p.bar.Grow(1, strconv.Itoa(nowAble))
	if recv == 0 {
		return
	}
	data := &utils.PingData{
		IP:       ip,
		Sended:   PingTimes,
		Received: recv,
		Delay:    totalDlay / time.Duration(recv),
	}
	p.buff.WriteString(ip.String() + "\n")
	p.appendIPData(data)
}
