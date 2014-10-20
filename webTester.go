package webtester

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"sort"
	"time"

	"github.com/juju/loggo"
)

var log = loggo.GetLogger("webtester")

func init() {
	log.SetLogLevel(loggo.INFO)
}

// CreateDataSample - Function type, which creates a new input data sample,
// which is passed on to the CreateNewRequest function to
// create a new request to the server using the created data
// sample as payload
type CreateDataSample func() interface{}

// CreateNewRequest - Function type that creates a new http request using
// the provided payload
type CreateNewRequest func(payload interface{}) (*http.Request, error)

// HandleResponse - Function type that handles the returned http response
type HandleResponse func(resp *http.Response) error

// ScenarioSamples - A struct type that servers as the basis for scenario definition.
// Each test scenario is composed of a sequence of samples, which define http requests
// fired of at certain points in time using predefined payloads
type ScenarioSample struct {
	Time    time.Duration `json:"time"`
	Payload interface{}   `json:"payload"`
}

type ScenarioExecutor struct {
}

type statusResponse struct {
	status   int
	err      error
	duration time.Duration
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func randomDuration(max int) int {
	return rand.Intn(max)
}

func sortedRandomSamples(len int, max int) []int {
	samples := make([]int, len)

	// create an array of random integers
	for i := 0; i < len; i++ {
		samples[i] = randomDuration(max)
	}
	sort.Ints(samples)

	return samples
}

func randomMoments(duration time.Duration, requestsPerSecond int) []time.Duration {

	var moments []time.Duration

	// first we create samples for a defined number of seconds (<duration)
	var t0 time.Duration
	for {
		t1 := t0 + time.Second
		if t1 > duration {
			break
		}

		ts := sortedRandomSamples(requestsPerSecond, 1000000000)

		mts := make([]time.Duration, requestsPerSecond)
		for i := 0; i < requestsPerSecond; i++ {
			mts[i] = t0 + time.Duration(ts[i])*time.Nanosecond
		}

		moments = append(moments, mts...)

		t0 = t1
	}

	// for the leftover fraction of a second we must add additional samples
	if t0 < duration {
		additionalSamples := int(math.Floor(float64(requestsPerSecond) * float64((duration - t0)) / float64(time.Second)))

		ts := sortedRandomSamples(additionalSamples, int(math.Floor(float64(1000000000*(duration-t0))/float64(time.Second))))

		mts := make([]time.Duration, additionalSamples)
		for i := 0; i < additionalSamples; i++ {
			mts[i] = t0 +
				time.Duration(ts[i])*time.Nanosecond
		}

		moments = append(moments, mts...)
	}
	return moments
}

// CreateNewTestScenario creates a new test scenario of length specified by the "duration" parameter.
// During the test the average load on the server will be as specified by the "requestsPerSecond" paramerer.
// For each request the input data is created by using the provided InputDataSampler and the resulting
// scenario is written to the file specified by the "scenarioFile" parameter.
func CreateNewTestScenario(duration time.Duration, requestsPerSecond int, ids CreateDataSample, scenarioFile string) {

	moments := randomMoments(duration, requestsPerSecond)

	samples := make([]ScenarioSample, len(moments))

	for i, t0 := range moments {

		s := ScenarioSample{t0, nil}
		if ids != nil {
			s.Payload = ids()
		}

		samples[i] = s
	}

	data, err := json.Marshal(samples)
	if err != nil {
		log.Errorf("Error marshalling input data to json :: %v", err)
	}

	err = ioutil.WriteFile(scenarioFile, data, 0644)
	if err != nil {
		log.Errorf("Error writing scenario to the target file :: %v", err)
	}
}

func executeRequest(resultChanel chan statusResponse, t0 time.Duration, client *http.Client, request *http.Request, hr HandleResponse) {
	if client == nil {
		log.Criticalf("Nil client.")
		return
	}
	if request == nil {
		log.Criticalf("Nil request.")
		return
	}

	timer := time.NewTimer(t0)

	<-timer.C

	var err error

	tStart := time.Now()
	response, err := client.Do(request)
	if err != nil {
		log.Errorf("Error sending a request via the specified client :: %v", err)
		resultChanel <- statusResponse{response.StatusCode, err, time.Since(tStart)}
		return
	}

	if hr != nil {
		err = hr(response)
		if err != nil {
			log.Errorf("Error handling the received response :: %v", response)
			resultChanel <- statusResponse{response.StatusCode, err, time.Since(tStart)}
			return
		}
	}

	resultChanel <- statusResponse{response.StatusCode, nil, time.Since(tStart)}
}

func ExecuteTestScenario(scenarioFile string, client *http.Client, cnr CreateNewRequest, hr HandleResponse) error {

	if client == nil {
		log.Criticalf("Nil client.")
		return errors.New("Nil client")
	}

	data, err := ioutil.ReadFile(scenarioFile)
	if err != nil {
		log.Errorf("Error reading scenario file :: %v", err)
		return err
	}

	var samples []ScenarioSample
	err = json.Unmarshal(data, &samples)
	if err != nil {
		log.Errorf("Error converting contents of scenario file to scenario samples :: %v", err)
		return err
	}

	resultChanel := make(chan statusResponse)

	for _, sample := range samples {
		request, err := cnr(sample.Payload)
		if err != nil {
			log.Errorf("Error creating a new http request from payload :: %v |-> %v", sample.Payload, err)
			return err
		}

		go executeRequest(resultChanel, sample.Time, client, request, hr)

	}

	var noSamples = len(samples)
	var totalDuration time.Duration
	for {
		s := <-resultChanel
		if s.err != nil {
			log.Warningf("Received error :: %v", err)
		}
		if s.status != 200 {
			log.Warningf("Received status code :: %v", s.status)
		}
		totalDuration = totalDuration + s.duration
		noSamples = noSamples - 1

		if noSamples <= 0 {
			break
		}
	}

	log.Infof("Average response time :: %v", time.Duration(math.Floor(float64(totalDuration)/float64(len(samples)))))
	return nil
}
