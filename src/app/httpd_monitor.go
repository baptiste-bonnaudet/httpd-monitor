package main

import (
	"container/ring"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"
	"github.com/joho/godotenv"
)

type request struct {
	method      string
	path        string
	httpVersion string
}

type logLine struct {
	ip        string
	clientid  string
	userid    string
	timestamp time.Time
	request   request
	status    int
	size      int
}

type stats struct {
	reqCounts *ring.Ring // ring buffer that contains request counts for the last alertLoopSeconds
	sections  map[string]int
	status    map[int]int
	alert     bool
	mux       sync.Mutex
}

const notifyLoopSeconds = 10
const alertLoopSeconds = 120
const topHitOutout = 5

var splitRegex = regexp.MustCompile("'.+'|\".+\"|\\[.+\\]|\\S+")
var apacheTimestampLayout = "[02/Jan/2006:15:04:05 -0700]"

func ParseLog(line string) logLine {
	log := logLine{}
	parts := splitRegex.FindAllString(line, -1)
	for k, v := range parts {
		v := strings.Replace(v, "\"", "", -1)
		if parts[k] != "-" {
			switch k {
			case 0:
				log.ip = v
			case 1:
				log.clientid = v
			case 2:
				log.userid = v
			case 3:
				t, err := time.Parse(apacheTimestampLayout, v)
				if err != nil {
					fmt.Println(err)
				}
				log.timestamp = t
			case 4:
				req := strings.Fields(v)
				log.request = request{method: req[0], path: req[1], httpVersion: req[2]}
			case 5:
				i, err := strconv.Atoi(v)
				if err != nil {
					fmt.Println(err)
				}
				log.status = i
			case 6:
				i, err := strconv.Atoi(v)
				if err != nil {
					fmt.Println(err)
				}
				log.size = i
			}
		}
	}
	return log
}

func UpdateStats(stats *stats, log logLine) {

	stats.mux.Lock()
	defer stats.mux.Unlock()

	if stats.reqCounts.Value == nil {
		stats.reqCounts.Value = 1
	} else {
		stats.reqCounts.Value = stats.reqCounts.Value.(int) + 1
	}

	stats.status[log.status]++
	section := strings.Split(log.request.path, "/")[1]
	stats.sections[section]++
}

func monitor(filePath string, stats *stats) {
	t, err := tail.TailFile(filePath, tail.Config{Follow: true})
	if err != nil {
		log.Fatal(err)
	}

	for line := range t.Lines {
		log := ParseLog(line.Text)
		// TODO: do not count entries older than alertLoopSeconds
		UpdateStats(stats, log)
	}
}

func AlertAndNotify(maxAvgMessages int, stats *stats) (message string) {

	//Aquire lock
	stats.mux.Lock()

	timestamp := time.Now()

	// sort Status
	type kvStatus struct {
		Key   int
		Value int
	}

	var topStatus []kvStatus
	for k, v := range stats.status {
		topStatus = append(topStatus, kvStatus{k, v})
	}

	sort.Slice(topStatus, func(i, j int) bool {
		return topStatus[i].Value > topStatus[j].Value
	})

	// Sort Sections
	type kvSections struct {
		Key   string
		Value int
	}

	var topSections []kvSections
	for k, v := range stats.sections {
		topSections = append(topSections, kvSections{k, v})
	}

	sort.Slice(topSections, func(i, j int) bool {
		return topSections[i].Value > topSections[j].Value
	})

	// Notify output
	message = "--------------------------------------------------------------\n"
	message += "[Notify]\n"
	message += fmt.Sprintf("Timestamp: %s \n", timestamp)
	message += fmt.Sprintf("Requests: %d \n", stats.reqCounts.Value)
	message += fmt.Sprintf("Top hit sections:\n")
	for k, v := range topSections {
		if k > topHitOutout {
			break
		}
		message += fmt.Sprintf("\t- %s: %d \n", v.Key, v.Value)
	}
	message += fmt.Sprintf("Top hit statuses:\n")
	for k, v := range topStatus {
		if k > topHitOutout {
			break
		}
		message += fmt.Sprintf("\t- %d: %d \n", v.Key, v.Value)
	}

	// Alert
	var totalCount int

	for i := 1; i <= stats.reqCounts.Len(); i++ {
		stats.reqCounts = stats.reqCounts.Next()
		if stats.reqCounts.Value != nil {
			totalCount += stats.reqCounts.Value.(int)
		}
	}

	if (totalCount > maxAvgMessages) && (stats.alert == false) {
		message += fmt.Sprintf("[Alert] Triggered - %d requests over the past %d seconds\n", totalCount, alertLoopSeconds)
		stats.alert = true
	} else if (totalCount <= maxAvgMessages) && (stats.alert == true) {
		message += fmt.Sprintf("[Alert] Recovered - %d requests over the past %d seconds\n", totalCount, alertLoopSeconds)
		stats.alert = false
	}

	if stats.alert == true {
		message += fmt.Sprintf("Alert Status: ON")
	} else {
		message += fmt.Sprintf("Alert Status: OFF")
	}

	// Cleanup
	stats.reqCounts = stats.reqCounts.Move(1) //slide the ring buffer 1 stop
	stats.reqCounts.Value = 0
	stats.sections = make(map[string]int)
	stats.status = make(map[int]int)

	// Release lock
	stats.mux.Unlock()

	return
}

func main() {

	godotenv.Load()

	accessLog := os.Getenv("ACCESS_LOG")
	maxAvgMessages, _ := strconv.Atoi(os.Getenv("MAX_AVERAGE_MESSAGES"))

	ring := ring.New(alertLoopSeconds / notifyLoopSeconds)
	for i := 0; i < ring.Len(); i++ {
		ring = ring.Next()
		ring.Value = 0
	}
	stats := stats{
		ring,
		make(map[string]int),
		make(map[int]int),
		false,
		sync.Mutex{},
	}

	// Monitoring subroutine
	go monitor(accessLog, &stats)

	// Main thread
	for {
		time.Sleep(notifyLoopSeconds * time.Second)
		fmt.Println(AlertAndNotify(maxAvgMessages, &stats))
	}
}
