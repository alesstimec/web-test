A simple web server testing framework based on fixed scenarios for measuring average response times.


** DEFINITIONS ** 

A scenario consists of a fixed sequence of ScenarioSamples, which are defined as follows:

type ScenarioSample struct {
	Time    time.Duration `json:"time"`
	Payload interface{}   `json:"payload"`
}

Each of these samplese is used to create an http Request using the CreateDataSample function type and sent to the server at a defined time.

// CreateDataSample - A user defined function, used to create new data samples. How the samples
// are created is entirely up to the user, but i recomment that these be randomly generated 
// samples
type CreateDataSample func() interface{}

// CreateNewRequest - A user defined function, which creates an http request based on the 
// input data sample provided by the CreateDataSample function
type CreateNewRequest func(payload interface{}) (*http.Request, error)

// HandleResponse - A user defined function, which is used for response handling, logging, etc.
type HandleResponse func(resp *http.Response) error


To create a new test scenario the CreateNewTestScenario should be used. This creates a test scenario of 
length specified by the "duration" parameter. During the test the average load of the server is specified by the number
of requests that are to be fired each second (specified by the "requestsPerSecond" parameter). For each request the input data is created by using the provided CreateDataSample function and the resulting scenario is written to the file specified by the "scenarioFile" parameter.

ExecuteTestScenario function executes a test scenario description stored in the "scenarioFile". Http client
must be provided, which is reused for all connections. In addition one must provide two fucntions: 1) CreateNewRequest
function which takes interface{} as input and produces a new http Request to the desired server. Input is read from
scenario definition file (json - only exported fields are available). 2) The HandleResponse function which is given
http Response and returns an error - in this function one may do erro handling, additional logging, etc.
