package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"errors"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
	"github.com/joho/godotenv"

)

const (
	maxConcurrentRequests = 2         // Controls concurrency
	rateLimitDelay        = 2 * time.Second // Delay between requests
	maxRetries            = 5
)

var osOpen = os.Open // default to actual os.Open

type JobListing struct {
	JobPosition    string `json:"job_position"`
	JobLink        string `json:"job_link"`
	JobID          string `json:"job_id"`
	CompanyName    string `json:"company_name"`
	CompanyProfile string `json:"company_profile"`
	JobLocation    string `json:"job_location"`
	JobPostingDate string `json:"job_posting_date"`
}

type JobDescription struct {
	JobPosition         string             `json:"job_position"`
	JobLocation         string             `json:"job_location"`
	CompanyName         string             `json:"company_name"`
	CompanyLinkedInID   string             `json:"company_linkedin_id"`
	JobPostingTime      string             `json:"job_posting_time"`
	JobDescription      string             `json:"job_description"`
	SeniorityLevel      string             `json:"Seniority_level"`
	EmploymentType      string             `json:"Employment_type"`
	JobFunction         string             `json:"Job_function"`
	Industries          string             `json:"Industries"`
	JobApplyLink        string             `json:"job_apply_link"`
	RecruiterDetails    []Recruiter        `json:"recruiter_details"`
	SimilarJobs         []SimilarJob       `json:"similar_jobs"`
	PeopleAlsoViewed    []SimilarJob       `json:"people_also_viewed"`
}

type Recruiter struct {
	RecruiterName  string `json:"recruiter_name"`
	RecruiterTitle string `json:"recruiter_title"`
}

type SimilarJob struct {
	JobPosition     string `json:"job_position"`
	JobCompany      string `json:"job_company"`
	JobLocation     string `json:"job_location"`
	JobPostingTime  string `json:"job_posting_time"`
	JobLink         string `json:"job_link"`
}

func main() {
	log.Println("Starting main function...")

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	log.Println(".env file loaded successfully")

	testMode := false
	var jobListings []JobListing

	if testMode {
		log.Println("Running in test mode")
		jobListings, err = getJobListingsTest()
		if err != nil {
			log.Fatalf("Error in getJobListingsTest: %v", err)
		}
		log.Printf("Loaded %d job listings from test JSON\n", len(jobListings))
	} else {
		log.Println("Running in production mode")
		jobListings, err = getJobListings()
		if err != nil {
			log.Fatalf("Error in getJobListings: %v", err)
		}
		log.Printf("Loaded %d job listings from API\n", len(jobListings))
	}

	log.Println("Processing job listings...")
	jobDescriptions := processJobListings(jobListings)
	log.Printf("Received %d job descriptions\n", len(jobDescriptions))

	for _, desc := range jobDescriptions {
		eval := getJobEvaluation(desc)
		fmt.Println(eval)
	}
	return
}

func getJobListingsTest() ([]JobListing, error) {
	log.Println("Opening test JSON file: myJSON.json")
	var jobListings []JobListing

	file, err := os.Open("myJSON.json")
	if err != nil {
		log.Printf("Failed to open test file: %v\n", err)
		return jobListings, err
	}
	defer file.Close()

	log.Println("Decoding test JSON data")
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&jobListings)
	if err != nil {
		log.Printf("Failed to decode test JSON: %v\n", err)
		return jobListings, err
	}

	return jobListings, nil
}

func getJobListings() ([]JobListing, error) {
	log.Println("Fetching job listings from API...")
	var allJobListings []JobListing

	apiKey := os.Getenv("SCRAPINGDOG_API_KEY")
	if apiKey == "" {
		return nil, errors.New("No API Key set in .env")
	}

	field := url.QueryEscape("Software Engineer")
	location := url.QueryEscape("Kansas City")
	geoid := "106142749"
	sortBy := "day"
	jobType := ""
	expLevel := ""
	workType := ""
	filterByCompany := ""

	const pageSize = 10
	page := 1

	for {
		log.Printf("Requesting page %d from API\n", page)
		url := fmt.Sprintf(
			"https://api.scrapingdog.com/linkedinjobs?api_key=%s&field=%s&geoid=%s&location=%s&page=%d&sort_by=%s&job_type=%s&exp_level=%s&work_type=%s&filter_by_company=%s",
			apiKey, field, geoid, location, page, sortBy, jobType, expLevel, workType, filterByCompany,
		)

		res, err := http.Get(url)
		if err != nil {
			log.Printf("Error fetching page %d: %v\n", page, err)
			return nil, err
		}
		defer res.Body.Close()

		var pageListings []JobListing
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(&pageListings)
		if err != nil {
			log.Printf("Error decoding page %d: %v\n", page, err)
			return nil, err
		}

		log.Printf("Fetched %d listings from page %d\n", len(pageListings), page)
		allJobListings = append(allJobListings, pageListings...)

		if len(pageListings) < pageSize {
			log.Println("Less than 10 listings returned — ending pagination")
			break
		}

		page++
	}

	return allJobListings, nil
}
func getJobDescriptionWithRetry(job JobListing) (JobDescription, error) {
	var desc JobDescription
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		desc, err = getJobDescription(job)
		if err == nil {
			return desc, nil
		}

		wait := time.Duration(attempt*2) * time.Second
		log.Printf("Retrying job %s (attempt %d/%d) after error: %v\n", job.JobID, attempt, maxRetries, err)
		time.Sleep(wait)
	}

	return desc, fmt.Errorf("failed after %d retries: %v", maxRetries, err)
}


func getJobDescription(job JobListing) (JobDescription, error) {
	log.Printf("Fetching description for JobID: %s (%s)\n", job.JobID, job.JobPosition)
	var desc JobDescription

	apiKey := os.Getenv("SCRAPINGDOG_API_KEY")
	if apiKey == "" {
		return desc, errors.New("No API Key set in .env")
	}

	if job.JobID == "" {
		log.Println("JobID is empty!")
		return desc, errors.New("Job link is empty")
	}

	apiURL := fmt.Sprintf("https://api.scrapingdog.com/linkedinjobs?api_key=%v&job_id=%v", apiKey, url.QueryEscape(job.JobID))

	const maxRetries = 5
	var resp *http.Response
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		resp, err = http.Get(apiURL)
		if err != nil {
			log.Printf("HTTP request failed for JobID %s: %v\n", job.JobID, err)
			return desc, err
		}

		if resp.StatusCode == http.StatusTooManyRequests { // 429
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			wait := time.Duration(attempt*2) * time.Second
			log.Printf("Rate limit hit for JobID %s: %s - %s. Retrying in %v...", job.JobID, resp.Status, string(bodyBytes), wait)
			time.Sleep(wait)
			continue
		} else if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			log.Printf("ScrapingDog error for JobID %s: %s - %s\n", job.JobID, resp.Status, string(bodyBytes))
			return desc, fmt.Errorf("ScrapingDog error: %s - %s", resp.Status, string(bodyBytes))
		}

		// Successful response, break retry loop
		break
	}

	if resp == nil {
		return desc, errors.New("Failed to get response from API after retries")
	}
	defer resp.Body.Close()

	log.Printf("Decoding job description for JobID: %s\n", job.JobID)
	decoder := json.NewDecoder(resp.Body)

	var descs []JobDescription
	err = decoder.Decode(&descs)
	if err != nil {
		log.Printf("Failed to decode job description array for JobID %s: %v\n", job.JobID, err)
		return desc, err
	}

	if len(descs) == 0 {
		err = errors.New("empty job description array")
		log.Printf("No job descriptions found for JobID %s\n", job.JobID)
		return desc, err
	}

	desc = descs[0]
	return desc, nil
}


type jobResult struct {
	desc JobDescription
	err  error
}

func processJobListings(jobListings []JobListing) []string {
	log.Println("Launching throttled goroutines for job descriptions")

	resultChan := make(chan jobResult)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxConcurrentRequests)
	ticker := time.NewTicker(rateLimitDelay)

	for _, job := range jobListings {
		wg.Add(1)
		go func(job JobListing) {
			defer wg.Done()

			semaphore <- struct{}{} // acquire slot
			<-ticker.C              // wait for rate limit delay

			desc, err := getJobDescriptionWithRetry(job)
			resultChan <- jobResult{desc: desc, err: err}

			<-semaphore // release slot
		}(job)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		ticker.Stop()
	}()

	return collectFormattedResults(resultChan)
}

func collectFormattedResults(resultChan <-chan jobResult) []string {
	var results []string

	log.Println("Collecting job descriptions from channel...")
	for res := range resultChan {
		if res.err != nil {
			log.Printf("Error occurred during description fetch: %v\n", res.err)
			continue
		}

		desc := res.desc
		log.Printf("Formatting job: %s at %s\n", desc.JobPosition, desc.CompanyName)
		formatted := fmt.Sprintf(
			`Title: %s
Company: %s
Location: %s
Posted: %s
Seniority Level: %s
Employment Type: %s
Job Function: %s
Industry: %s
Apply Link: %s
Description: %s
---`,
			desc.JobPosition,
			desc.CompanyName,
			desc.JobLocation,
			desc.JobPostingTime,
			desc.SeniorityLevel,
			desc.EmploymentType,
			desc.JobFunction,
			desc.Industries,
			desc.JobApplyLink,
			desc.JobDescription,
		)

		results = append(results, formatted)
	}

	log.Println("Finished formatting job descriptions")
	return results
}
