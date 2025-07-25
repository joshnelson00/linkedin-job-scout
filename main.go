package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	// "net/http" // Uncomment for production
	// "net/url"  // Uncomment for production

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

	// ========== PRODUCTION CONFIG (uncomment to use API) ==========
	/*
		apiKey := os.Getenv("SCRAPINGDOG_API_KEY")
		if apiKey == "" {
			log.Fatal("No API Key set in .env")
		}

		// === Customize API Query Parameters Below ===
		field := url.QueryEscape("Software Engineer") // Job title
		location := url.QueryEscape("Kansas City")    // Optional: leave empty for all locations
		geoid := "106142749"                          // LinkedIn GeoID (Kansas City Metro)
		page := "1"                                   // Pagination
		sortBy := "day"                               // "relevance" or "day"
		jobType := ""                                 // e.g., "F" for Full-time
		expLevel := ""                                // e.g., "1" for Entry Level
		workType := ""                                // e.g., "1" for On-site, "2" for Remote
		filterByCompany := ""                         // Company name or ID (optional)

		url := fmt.Sprintf(
			"https://api.scrapingdog.com/linkedinjobs?api_key=%s&field=%s&geoid=%s&location=%s&page=%s&sort_by=%s&job_type=%s&exp_level=%s&work_type=%s&filter_by_company=%s",
			apiKey, field, geoid, location, page, sortBy, jobType, expLevel, workType, filterByCompany,
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
	*/

	// ========== TEMPORARY LOCAL TESTING ==========
	file, err := os.Open("myJSON.json")
	if err != nil {
		log.Fatalf("Failed to open local JSON file: %v", err)
	}
	defer file.Close()

	var JobListings []JobListing
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&JobListings)
	if err != nil {
		log.Fatalf("Failed to decode local JSON: %v", err)
	}

	// Pretty print the response
	jobsJSON, err := json.MarshalIndent(JobListings, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(jobsJSON))
}
