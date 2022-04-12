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
	pendingApprovals   *prometheus.Desc
	trackedImagesTotal *prometheus.Desc
	registriesTotal    *prometheus.Desc
	namespacesTotal    *prometheus.Desc
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
type trackedItem struct {
	Image        string `json:"image"`
	Trigger      string `json:"trigger"`
	PollSchedule string `json:"pollSchedule"`
	Provider     string `json:"provider"`
	Namespace    string `json:"namespace"`
	Policy       string `json:"policy"`
	Registry     string `json:"registry"`
}

type trackedStats struct {
	imagesTotal     int
	namespacesTotal int
	registriesTotal int
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
		pendingApprovals:   prometheus.NewDesc("keel_pending_approvals_total", "Updates pending in Keel", nil, nil),
		trackedImagesTotal: prometheus.NewDesc("keel_tracked_images_total", "Total number of images tracked by Keel", nil, nil),
		registriesTotal:    prometheus.NewDesc("keel_registries_total", "Images are tracked across these many registries", nil, nil),
		namespacesTotal:    prometheus.NewDesc("keel_namespaces_total", "Images are tracked across these many namespaces", nil, nil),
	}
}

//Each and every collector must implement the Describe function.
//It essentially writes all descriptors to the prometheus desc channel.
func (collector *KeelCollector) Describe(ch chan<- *prometheus.Desc) {

	//Update this section with the each metric you create for a given collector
	ch <- collector.pendingApprovals
	ch <- collector.trackedImagesTotal
	ch <- collector.registriesTotal
	ch <- collector.namespacesTotal
}

//Collect implements required collect function for all promehteus collectors
func (collector *KeelCollector) Collect(ch chan<- prometheus.Metric) {

	//Write latest value for each metric in the prometheus metric channel.
	//Note that you can pass CounterValue, GaugeValue, or UntypedValue types here.
	approvals := prometheus.MustNewConstMetric(collector.pendingApprovals, prometheus.GaugeValue, getPendingApprovals())
	ch <- approvals

	trackedStatsValue, err := getTrackedStats()
	if err != nil {
		fmt.Println("Error %s", err)
		return
	}
	images := prometheus.MustNewConstMetric(collector.trackedImagesTotal, prometheus.CounterValue, float64(trackedStatsValue.imagesTotal))
	ch <- images
	namespaces := prometheus.MustNewConstMetric(collector.namespacesTotal, prometheus.CounterValue, float64(trackedStatsValue.namespacesTotal))
	ch <- namespaces
	registries := prometheus.MustNewConstMetric(collector.registriesTotal, prometheus.CounterValue, float64(trackedStatsValue.registriesTotal))
	ch <- registries

}

func getTrackedStats() (trackedStats, error) {
	var body, err = call("/v1/tracked", "GET")
	if err != nil {
		return trackedStats{}, err
	}
	var trackedItems []trackedItem
	if err := json.Unmarshal(body, &trackedItems); err != nil { // Parse []byte to go struct pointer
		return trackedStats{}, err
	}

	var stats trackedStats

	var namespaces = Set{}
	var registries = Set{}

	for _, v := range trackedItems {
		stats.imagesTotal++
		namespaces.add(v.Namespace)
		registries.add(v.Registry)
	}

	stats.namespacesTotal = len(namespaces)
	stats.registriesTotal = len(registries)
	return stats, err
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
