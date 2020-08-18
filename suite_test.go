package enqueuestomp_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	check "gopkg.in/check.v1"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { check.TestingT(t) }

type EnqueueStompSuite struct{}

var _ = check.Suite(&EnqueueStompSuite{})

func (s *EnqueueStompSuite) TearDownTest(c *check.C) {
	time.Sleep(10 * time.Millisecond)
}

type jolokia struct {
	addr     string
	username string
	passwd   string
	c        *check.C
}

func (j *jolokia) GetStats(destinationType string, destinationName string) {
	url := j.makeURLStats(destinationType, destinationName)
	print(url)
	j.doRequest(url)
}

func (j *jolokia) doRequest(url string) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil) // nolint:noctx
	j.checkError(err)

	req.SetBasicAuth(j.username, j.passwd)
	resp, err := client.Do(req)
	j.checkError(err)
	defer resp.Body.Close()

	bodyText, err := ioutil.ReadAll(resp.Body)
	j.checkError(err)

	fmt.Printf("bodyText %s", string(bodyText))
	// json.Marshal()
}

func (j *jolokia) checkError(err error) {
	if err != nil {
		j.c.Fatal(err)
	}
}

func (j *jolokia) makeURLStats(destinationType string, destinationName string) string {
	path := fmt.Sprintf(
		"/read/org.apache.activemq:type=Broker,brokerName=localhost,destinationType=%s,destinationName=%s",
		destinationType, destinationName,
	)
	return j.makeURLJolokia(path)
}

func (j *jolokia) makeURLJolokia(path string) string {
	return fmt.Sprintf("%s/api/jolokia%s", j.addr, path)
}
