package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alesstimec/web-tester"
)

func CreateNewInputData() interface{} {
	return 1
}

func CreateNewRequest(data interface{}) (*http.Request, error) {
	b, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error converting data to json :: %v", err)
		return nil, err
	}
	reader := bytes.NewReader(b)

	req, err := http.NewRequest("POST", "http://localhost:8080", reader)
	return req, err
}

func main() {

	if _, err := os.Stat("test_scenario.tsc"); os.IsNotExist(err) {
		webtester.CreateNewTestScenario(time.Duration(10)*time.Second, 1, CreateNewInputData, "test_scenario.tsc")
	} else {
		fmt.Println("Reusing existing scenario file [test_scenario.tsc]")
	}

	client := &http.Client{}

	err := webtester.ExecuteTestScenario("test_scenario.tsc", client, CreateNewRequest, nil)
	if err != nil {
		fmt.Println("An error has occured :: %v", err)
	}

}
