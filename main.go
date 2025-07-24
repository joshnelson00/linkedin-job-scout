package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"net/url"
	"github.com/joho/godotenv"
)

func main() {
	type JobListing struct {
		JobPosition    string `json:"job_position"`
		JobLink        string `json:"job_link"`
		JobID          string `json:"job_id"`
		CompanyName    string `json:"company_name"`
		CompanyProfile string `json:"company_profile"`
		JobLocation    string `json:"job_location"`
		JobPostingDate string `json:"job_posting_date"`
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	apiKey := os.Getenv("SCRAPINGDOG_API_KEY")
	if apiKey == "" {
		log.Fatal("No API Key set in .env")
	}

	field := url.QueryEscape("Software Engineer")
	if field == "" {
		log.Fatal("No Field set in URL")
	}

	url := fmt.Sprintf(
		"https://api.scrapingdog.com/linkedinjobs?api_key=%v&field=%v&geoid=106142749&location=&page=1&sort_by=day&job_type=&exp_level=&work_type=&filter_by_company=",
		apiKey,
		field,
	)

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	var JobListings []JobListing
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&JobListings)
	if err != nil {
		log.Fatal(err)
	}

	// Pretty print the response
	jobsJSON, err := json.MarshalIndent(JobListings, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jobsJSON))
}
