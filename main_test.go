// main_test.go
package main

import (
	"strings"
	"testing"
)

func TestCollectFormattedResults(t *testing.T) {
	// Create a channel and feed it test jobResults
	resultChan := make(chan jobResult, 2)

	resultChan <- jobResult{
		desc: JobDescription{
			JobPosition:    "Software Engineer",
			CompanyName:    "Tech Corp",
			JobLocation:    "New York, NY",
			JobPostingTime: "2025-08-10",
			SeniorityLevel: "Internship",
			EmploymentType: "Full-time",
			JobFunction:    "Engineering",
			Industries:     "Software",
			JobApplyLink:   "http://apply.here",
			JobDescription: "Build cool things.",
		},
		err: nil,
	}

	// Second job has an error and should be skipped
	resultChan <- jobResult{
		desc: JobDescription{},
		err:  assertError{}, // custom dummy error
	}

	close(resultChan)

	results := collectFormattedResults(resultChan)

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	// Basic content checks
	out := results[0]
	expectedParts := []string{
		"Title: Software Engineer",
		"Company: Tech Corp",
		"Location: New York, NY",
		"Apply Link: http://apply.here",
		"Description: Build cool things.",
	}
	for _, part := range expectedParts {
		if !strings.Contains(out, part) {
			t.Errorf("Output missing expected part: %q", part)
		}
	}
}

// Dummy error type for testing
type assertError struct{}

func (e assertError) Error() string { return "dummy error" }
