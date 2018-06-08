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
	"time"

	"github.com/davecgh/go-spew/spew"
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
}

const notifyLoopSeconds = 4
const alertLoopSeconds = 12

var splitRegex = regexp.MustCompile("'.+'|\".+\"|\\[.+\\]|\\S+")
var apacheTimestampLayout = "[02/Jan/2006:15:04:05 -0700]"

func parseLog(line string) logLine {
	log := logLine{}
	parts := splitRegex.FindAllString(line, -1)
	for k, v := range parts {
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

func updateStats(stats *stats, log logLine) {
	if stats.reqCounts.Value == nil {
		stats.reqCounts.Value = 1
	} else {
		stats.reqCounts.Value = stats.reqCounts.Value.(int) + 1
	}

	stats.status[log.status] = stats.status[log.status] + 1
	section := strings.Split(log.request.path, "/")[1]
	stats.sections[section] = stats.sections[section] + 1
}

func monitor(filePath string, stats *stats) {
	t, err := tail.TailFile(filePath, tail.Config{Follow: true})
	if err != nil {
		log.Fatal(err)
	}

	for line := range t.Lines {
		log := parseLog(line.Text)
		spew.Dump(log)
		updateStats(stats, log)
	}
}

func alertAndNotify(stats *stats) {

	for {
		time.Sleep(notifyLoopSeconds * time.Second)
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

		fmt.Println("--------------------------------------------------------------")
		fmt.Printf("[Notify]\n")
		fmt.Printf("Timestamp: %s \n", timestamp)
		fmt.Printf("Requests: %d \n", stats.reqCounts.Value)
		fmt.Printf("Top hit sections:\n")
		for k, v := range topSections {
			if k > 5 {
				break
			}
			fmt.Printf("\t- %s: %d \n", v.Key, v.Value)
		}
		fmt.Printf("Top hit status:\n")
		for k, v := range topStatus {
			if k > 5 {
				break
			}
			fmt.Printf("\t- %d: %d \n", v.Key, v.Value)
		}

		// Cleanup
		stats.reqCounts = stats.reqCounts.Move(1) //slide the ring buffer 1 stop
		stats.reqCounts.Value = 0
		stats.sections = make(map[string]int)
		stats.status = make(map[int]int)

		// Alert

	}
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	accessLog := os.Getenv("ACCESS_LOG")

	stats := stats{
		ring.New(alertLoopSeconds / notifyLoopSeconds),
		make(map[string]int),
		make(map[int]int),
	}

	go monitor(accessLog, &stats)
	go alertAndNotify(&stats)

	for {
		// main thread
	}
}
