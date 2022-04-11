package collector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

//Define a struct for you collector that contains pointers
//to prometheus descriptors for each metric you wish to expose.
//Note you can also include fields of other types if they provide utility
//but we just won't be exposing them as metrics.
type KeelCollector struct {
	pendingApprovals *prometheus.Desc
}
type approval struct {
	ID         string `json:"id"`
	Archived   bool   `json:"archived"`
	Provider   string `json:"provider"`
	Identifier string `json:"identifier"`
	Event      struct {
		Repository struct {
			Host   string `json:"host"`
			Name   string `json:"name"`
			Tag    string `json:"tag"`
			Digest string `json:"digest"`
		} `json:"repository"`
		CreatedAt   time.Time `json:"createdAt"`
		TriggerName string    `json:"triggerName"`
	} `json:"event"`
	Message        string      `json:"message"`
	CurrentVersion string      `json:"currentVersion"`
	NewVersion     string      `json:"newVersion"`
	Digest         string      `json:"digest"`
	VotesRequired  int         `json:"votesRequired"`
	VotesReceived  int         `json:"votesReceived"`
	Voters         interface{} `json:"voters"`
	Rejected       bool        `json:"rejected"`
	Deadline       time.Time   `json:"deadline"`
	CreatedAt      time.Time   `json:"createdAt"`
	UpdatedAt      time.Time   `json:"updatedAt"`
}

type approvals []struct {
	approval
}

var (
	keelUrl  string
	keelUser string
	keelPass string
)

//You must create a constructor for you collector that
//initializes every descriptor and returns a pointer to the collector
func NewKeelCollector(url string, user string, pass string) *KeelCollector {
	keelUrl = url
	keelUser = user
	keelPass = pass
	return &KeelCollector{
		pendingApprovals: prometheus.NewDesc("keel_pending_approvals",
			"Updates pending in Keel",
			nil, nil,
		),
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *KeelCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.pendingApprovals
}

//Collect implements required collect function for all promehteus collectors
func (collector *KeelCollector) Collect(ch chan<- prometheus.Metric) {

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	m1 := prometheus.MustNewConstMetric(collector.pendingApprovals, prometheus.GaugeValue, getPendingApprovals())
	m1 = prometheus.NewMetricWithTimestamp(time.Now().Add(-time.Hour), m1)
	ch <- m1
}

func getPendingApprovals() float64 {
	var body, err = call("/v1/approvals", "GET")
	if err != nil {
		fmt.Println(err)
	}

	var approvals approvals
	if err := json.Unmarshal(body, &approvals); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	fmt.Println(PrettyPrint(approvals))
	var numberOfPendingApprovals int
	for _, v := range approvals {
		if !v.Archived {
			numberOfPendingApprovals++
		}
	}
	return float64(numberOfPendingApprovals)
}

func call(path, method string) ([]byte, error) {
	client := &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	req, err := http.NewRequest(method, keelUrl+path, nil)
	if err != nil {
		return nil, fmt.Errorf("Got error %s", err.Error())
	}
	req.SetBasicAuth(keelUser, keelPass)
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Got error %s", err.Error())
	}
	body, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	return body, err
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	return string(s)
}
