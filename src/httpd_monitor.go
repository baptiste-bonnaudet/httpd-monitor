package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
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
	reqCount int
	sections map[string]int
	status   map[int]int
}

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
	stats.reqCount = stats.reqCount + 1
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

func updateAndNotify(stats *stats) {
	for {
		time.Sleep(2 * time.Second)
		spew.Dump(stats)
		timestamp := time.Now()

		fmt.Printf("[Notify]\n")
		fmt.Printf("Timestamp: %s \n", timestamp)
		fmt.Printf("Requests: %d \n", stats.reqCount)
		fmt.Printf("Top hit sections:\n")
		fmt.Printf("Top hit status:\n")
		// for k, _ := range stats.status {
		// 	if k > 5 {
		// 		break
		// 	}
		// 	fmt.Printf("\t-\n")

		// }
		// type notify struct {
		// 	time     time.Time `json:"time"`
		// 	reqCount int       `json:"request_count"`
		// }

		// log, _ := json.Marshal(&notify{time: timestamp, reqCount: stats.reqCount})
		// spew.Dump(notify{time: timestamp, reqCount: stats.reqCount})
		// fmt.Println(log)
	}
	// (main.logLine) {
	// 	ip: (string) (len=10) "172.23.0.1",
	// 	clientid: (string) "",
	// 	userid: (string) "",
	// 	timestamp: (time.Time) 2018-06-07 21:46:36 +0000 +0000,
	// 	request: (main.request) {
	// 	 method: (string) (len=5) "\"POST",
	// 	 path: (string) (len=25) "/v1/auth/token/renew-self",
	// 	 httpVersion: (string) (len=9) "HTTP/1.1\""
	// 	},
	// 	status: (int) 404,
	// 	size: (int) 222
	//  }
}

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	accessLog := os.Getenv("ACCESS_LOG")

	stats := stats{0, make(map[string]int), make(map[int]int)}

	go monitor(accessLog, &stats)
	go updateAndNotify(&stats)

	for {
		// main thread
	}
}
