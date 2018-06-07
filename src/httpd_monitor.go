//127.0.0.1 user-identifier frank [10/Oct/2000:13:55:36 -0700] "GET /apache_pb.gif HTTP/1.0" 200 2326

// A "-" in a field indicates missing data.

// 127.0.0.1 is the IP address of the client (remote host) which made the request to the server.
// user-identifier is the RFC 1413 identity of the client.
// frank is the userid of the person requesting the document.
// [10/Oct/2000:13:55:36 -0700] is the date, time, and time zone that the request was received, by default in strftime format %d/%b/%Y:%H:%M:%S %z.
// "GET /apache_pb.gif HTTP/1.0" is the request line from the client. The method GET, /apache_pb.gif the resource requested, and HTTP/1.0 the HTTP protocol.
// 200 is the HTTP status code returned to the client. 2xx is a successful response, 3xx a redirection, 4xx a client error, and 5xx a server error.
// 2326 is the size of the object returned to the client, measured in bytes.

// Display stats every 10s about the traffic during those 10s:
// the sections of the web site with the most hits,
// as well as interesting summary statistics on the traffic as a whole.
// A section is defined as being what's before the second '/' in the path.
// For example, the section for "http://my.site.com/pages/create” is
// "http://my.site.com/pages".

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

var splitRegex = regexp.MustCompile("'.+'|\".+\"|\\[.+\\]|\\S+")
var apacheTimestampLayout = "[02/Jan/2006:15:04:05 -0700]"

func parseLog(line string) {
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
	spew.Dump(log)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	accessLog := os.Getenv("ACCESS_LOG")
	fmt.Println(accessLog)

	t, err := tail.TailFile(accessLog, tail.Config{Follow: true})
	for line := range t.Lines {
		parseLog(line.Text)
	}
}
