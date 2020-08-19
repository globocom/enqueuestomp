/*
* enqueuestomp
*
* MIT License
*
* Copyright (c) 2020 Globo.com
 */

package enqueuestomp_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	golog "log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/globocom/enqueuestomp"
	check "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	check.TestingT(t)
}

type EnqueueStompSuite struct {
	c *check.C
	j *jolokia
}

var _ = check.Suite(&EnqueueStompSuite{})

func (s *EnqueueStompSuite) SetUpSuite(c *check.C) {
	s.c = c
	s.j = &jolokia{
		addr:     "http://localhost:8161",
		username: "admin",
		passwd:   "admin",
		c:        c,
	}
}

func (s *EnqueueStompSuite) SetUpTest(c *check.C) {
	s.j.Delete(enqueuestomp.DestinationTypeQueue, queueName)
	s.j.Delete(enqueuestomp.DestinationTypeTopic, topicName)
}

func (s *EnqueueStompSuite) TearDownTest(c *check.C) {
	s.deleteFile(queueWriteOutputPath)
	s.deleteFile(topicWriteOutputPath)
	time.Sleep(20 * time.Millisecond)
}

func (s *EnqueueStompSuite) waitQueueSize(enqueue *enqueuestomp.EnqueueStomp) {
	for enqueue.QueueSize() > 0 {
		time.Sleep(300 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
}

func (s *EnqueueStompSuite) deleteFile(filePath string) {
	_ = os.Remove(filePath)
}

func (s *EnqueueStompSuite) countFileLines(filePath string) int {
	file, err := os.Open(filePath)
	s.c.Assert(err, check.IsNil)

	fileScanner := bufio.NewScanner(file)
	lineCount := 0
	for fileScanner.Scan() {
		lineCount++
	}
	return lineCount
}

//  https://gist.github.com/yashpatil/de7437522bfccfeee4cb#retrieve-all-attributes-2
type jolokia struct {
	addr     string
	username string
	passwd   string
	c        *check.C
}

func (j *jolokia) StatQueue(queueName string, attribute string) string {
	return j._stat(enqueuestomp.DestinationTypeQueue, queueName, attribute)
}

func (j *jolokia) StatTopic(topicName string, attribute string) string {
	return j._stat(enqueuestomp.DestinationTypeTopic, topicName, attribute)
}

func (j *jolokia) _stat(destinationType string, destinationName string, attribute string) string {
	url := j.makeURLStats(destinationType, destinationName, attribute)
	// println("url :%s", url)
	body := j.doRequest(url)

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		j.c.Assert(err, check.IsNil)
	}

	val := data["value"]
	switch val.(type) { // nolint
	case float64:
		return strconv.Itoa(int(val.(float64)))
	case int:
		return strconv.Itoa(val.(int))
	default:
		return fmt.Sprintf("%v", val)
	}
}

func (j *jolokia) Delete(destinationType string, destinationName string) {
	url := j.makeURLDelete(destinationType, destinationName)
	_ = j.doRequest(url)
}

func (j *jolokia) makeURLDelete(destinationType string, destinationName string) string {
	path := fmt.Sprintf(
		"/exec/org.apache.activemq:type=Broker,brokerName=localhost/remove%s(java.lang.String)/%s",
		strings.Title(destinationType), destinationName,
	)
	return j.makeURLJolokia(path)
}

func (j *jolokia) makeURLStats(destinationType string, destinationName string, attribute string) string {
	path := fmt.Sprintf(
		"/read/org.apache.activemq:type=Broker,brokerName=localhost,destinationType=%s,destinationName=%s/%s",
		strings.Title(destinationType), destinationName, attribute,
	)
	return j.makeURLJolokia(path)
}

func (j *jolokia) makeURLJolokia(path string) string {
	return fmt.Sprintf("%s/api/jolokia%s", j.addr, path)
}

func (j *jolokia) doRequest(url string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil) // nolint:noctx
	j.c.Assert(err, check.IsNil)

	req.SetBasicAuth(j.username, j.passwd)
	resp, err := client.Do(req)
	j.c.Assert(err, check.IsNil)

	defer resp.Body.Close()
	if resp.StatusCode > 399 {
		j.c.Errorf("jokia status code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	j.c.Assert(err, check.IsNil)

	return body
}

// increase resources limitation.
func changeMaxULimit(ulimit int) error {
	var uLimit syscall.Rlimit

	if err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &uLimit); err != nil {
		golog.Printf("error on get ulimit: %s", err)
		return err
	}

	if uLimit.Cur != uLimit.Max {
		golog.Printf("Current Ulimit value %d", uLimit.Cur)
		uLimit.Cur = uint64(ulimit)
		golog.Printf("changed Ulimit to %d", uLimit.Cur)

		if err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &uLimit); err != nil {
			golog.Printf("error on set ulimit : %s", err)
			return err
		}
	}
	return nil
}
