package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Month      string `json:"month"`
	Num        int    `json:"num"`
	Link       string `json:"link"`
	Year       string `json:"year"`
	News       string `json:"news"`
	SafeTitle  string `json:"safe_title"`
	Transcript string `json:"transcript"`
	Alt        string `json:"alt"`
	Img        string `json:"img"`
	Title      string `json:"title"`
	Day        string `json:"day"`
}

const Url = "https://xkcd.com"

func main() {

	// n := 223
	// result, err := fetch(n)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%v\n", result)

	//alocate jobs
	noOfJobs := 3000

	go allocateJobs(noOfJobs)

	done := make(chan bool)

	go getResult(done)

	//create worker pool
	noOfWorkers := 100
	createWorkerPool(noOfWorkers)

	<-done

	data, err := json.MarshalIndent(resultCollection, "", "	")

	if err != nil {
		log.Fatal("JSON error: ", err)
	}

	err = writeToFile(data)

	if err != nil {
		log.Fatal(err)
	}

}

func writeToFile(data []byte) error {
	f, err := os.Create("test.json")
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func fetch(n int) (*Result, error) {

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	url := strings.Join([]string{Url, fmt.Sprintf("%d", n), "info.0.json"}, "/")

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("HTTP Request: %v", err)
	}

	resp, err := client.Do(req)

	if err != nil {
		return nil, fmt.Errorf("HTTP Error: %v", err)
	}

	var data Result

	if resp.StatusCode != http.StatusOK {
		data = Result{}
	} else {
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return nil, fmt.Errorf("json err: %v", err)
		}
	}

	resp.Body.Close()

	return &data, nil
}

type Job struct {
	number int
}

var jobs = make(chan Job, 500)
var results = make(chan Result, 500)
var resultCollection []Result

func allocateJobs(noOfJobs int) {
	for i := 0; i <= noOfJobs; i++ {
		jobs <- Job{i + 1}
	}
	close(jobs)
}

func worker(wg *sync.WaitGroup) {
	for job := range jobs {
		result, err := fetch(job.number)
		if err != nil {
			log.Printf("Error in fetching: %v\n", err)
		}
		results <- *result
	}
	wg.Done()
}

func createWorkerPool(noOfWorkers int) {
	var wg sync.WaitGroup
	for i := 0; i <= noOfWorkers; i++ {
		wg.Add(1)
		go worker(&wg)
	}
	wg.Wait()
	close(results)
}

func getResult(done chan bool) {
	for result := range results {
		if result.Num != 0 {
			fmt.Printf("Retrieving issue #%d\n", result.Num)
			resultCollection = append(resultCollection, result)
		}
	}
	done <- true
}
