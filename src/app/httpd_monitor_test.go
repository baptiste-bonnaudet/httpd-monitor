package main

import (
	"container/ring"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"
)

func TestParseLog(t *testing.T) {

	layout := "2006-01-02 15:04:05 -0700"
	time, err := time.Parse(layout, "2018-06-07 15:28:25 +0000")

	if err != nil {
		fmt.Println(err)
	}
	a := logLine{
		ip:        "172.23.0.1",
		clientid:  "",
		userid:    "",
		timestamp: time,
		request:   request{method: "HEAD", path: "/test/test", httpVersion: "HTTP/1.1"},
		status:    404,
		size:      0,
	}

	b := ParseLog("172.23.0.1 - - [07/Jun/2018:15:28:25 +0000] \"HEAD /test/test HTTP/1.1\" 404 -")

	if reflect.DeepEqual(a, b) != true {
		t.Fail()
	}
}

func TestAlertAndNotify(t *testing.T) {

	maxAvgMessages := 55
	items := ring.New(10)
	for i := 0; i <= items.Len(); i++ {
		items = items.Next()
		items.Value = 5
	}

	s := stats{
		items,
		make(map[string]int),
		make(map[int]int),
		false,
		sync.Mutex{},
	}

	AlertAndNotify(maxAvgMessages, &s)
	if s.alert == true {
		t.Fail()
	}

	maxAvgMessages = 45
	items = ring.New(10)
	for i := 0; i <= items.Len(); i++ {
		items = items.Next()
		items.Value = 5
	}

	s = stats{
		items,
		make(map[string]int),
		make(map[int]int),
		false,
		sync.Mutex{},
	}

	AlertAndNotify(maxAvgMessages, &s)
	if s.alert == false {
		t.Fail()
	}
}
