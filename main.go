package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"errors"
	// "net/http" // Uncomment for production
	// "net/url"  // Uncomment for production

	"github.com/joho/godotenv"
)

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
	PeopleAlsoViewed    []SimilarJob       `json:"people_also_viewed"` // Same structure as SimilarJobs
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// ========== PRODUCTION CONFIG (uncomment to use API) ==========
	/*
		err = getJobListings()
		if err != nil {
			log.Fatalf("Error in getJobListings: %v", err)
		}
	*/
	jobListings, err := getJobListingsTest()
	if err != nil {
		log.Fatalf("Error in getJobListingsTest: %v", err)
	}

	// Create channel to collect JobDescriptions
	descChan := make(chan JobDescription)
	errChan := make(chan error)
	var wg sync.WaitGroup

	for _, job := range jobListings {
		wg.Add(1)
		go func(job JobListing) {
			defer wg.Done()
			desc, err := getJobDescription(job)
			if err != nil {
				errChan <- err
				return
			}
			descChan <- desc
		}(job)
	}

	// Close channels when done
	go func() {
		wg.Wait()
		close(descChan)
		close(errChan)
	}()

	// Collect results
	for {
		select {
		case desc, ok := <-descChan:
			if !ok {
				descChan = nil
			} else {
				// Handle description (print/store)
				fmt.Printf("Job: %s at %s\n", desc.JobPosition, desc.CompanyName)
			}
		case err, ok := <-errChan:
			if !ok {
				errChan = nil
			} else {
				log.Printf("Error: %v", err)
			}
		}
		if descChan == nil && errChan == nil {
			break
		}
	}
}


func getJobListingsTest() error {
	var jobListings []JobListing

	file, err := os.Open("myJSON.json")
	if err != nil {
		return jobListings, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&jobListings)
	if err != nil {
		return joblistings, err
	}

	return jobListings, nil
}

func getJobListings() ([]JobListing, error) {
	var allJobListings []JobListing

	apiKey := os.Getenv("SCRAPINGDOG_API_KEY")
	if apiKey == "" {
		return nil, errors.New("No API Key set in .env")
	}

	// === Static query params ===
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
		// Build URL for the current page
		url := fmt.Sprintf(
			"https://api.scrapingdog.com/linkedinjobs?api_key=%s&field=%s&geoid=%s&location=%s&page=%d&sort_by=%s&job_type=%s&exp_level=%s&work_type=%s&filter_by_company=%s",
			apiKey, field, geoid, location, page, sortBy, jobType, expLevel, workType, filterByCompany,
		)

		// Fetch this page
		res, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("error fetching page %d: %w", page, err)
		}
		defer res.Body.Close()

		var pageListings []JobListing
		decoder := json.NewDecoder(res.Body)
		err = decoder.Decode(&pageListings)
		if err != nil {
			return nil, fmt.Errorf("error decoding page %d: %w", page, err)
		}

		// Append to allJobListings Slice
		allJobListings = append(allJobListings, pageListings...)

		// Stop when fewer than 10 listings returned
		if len(pageListings) < pageSize {
			break
		}

		page++ // next page
	}

	return allJobListings, nil
}

func processJobListings(jobListings []JobListing) {
	descChan := make(chan JobDescription)
	errChan := make(chan error)
	var wg sync.WaitGroup

	for _, job := range jobListings {
		wg.Add(1)
		go func(job JobListing) {
			defer wg.Done()
			desc, err := getJobDescription(job)
			if err != nil {
				errChan <- err
				return
			}
			descChan <- desc
		}(job)
	}

	// Close channels after all jobs are processed
	go func() {
		wg.Wait()
		close(descChan)
		close(errChan)
	}()

	// Collect results
	collectResults(descChan, errChan)
}

func collectResults(descChan <-chan JobDescription, errChan <-chan error) {
	for descChan != nil || errChan != nil {
		select {
		case desc, ok := <-descChan:
			if !ok {
				descChan = nil
				continue
			}
			fmt.Printf("Job: %s at %s\n", desc.JobPosition, desc.CompanyName)

		case err, ok := <-errChan:
			if !ok {
				errChan = nil
				continue
			}
			log.Printf("Error: %v", err)
		}
	}
}

