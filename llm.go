// llm.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Request struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Stream      bool      `json:"stream"`
	Temperature float64   `json:"temperature,omitempty"` // <-- Add this
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Response struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Message            Message   `json:"message"`
	Done               bool      `json:"done"`
	TotalDuration      int64     `json:"total_duration"`
	LoadDuration       int       `json:"load_duration"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
}

const defaultOllamaURL = "http://localhost:11434/api/chat"

const systemInstruction = `You are an expert career advisor and resume evaluator. You return strict but accurate feedback with practical suggestions.`

type Evaluation struct {
	Score int
	Text  string
}

func getJobEvaluations(jobDescs []string) error {
	fmt.Println("🔍 Starting getJobEvaluations")

	resumeBytes, err := os.ReadFile("resume.txt")
	if err != nil {
		return err
	}
	resumeContent := string(resumeBytes)

	outputFile := "LinkedinEvaluations.html"

	fmt.Printf("🗑️  Resetting output file: %s\n", outputFile)

	modelName := os.Getenv("OLLAMA_MODEL")
	if modelName == "" {
		modelName = "gemma3:1b" // Model preference here
	}
	temperature := 0.3 // default value
	if tStr := os.Getenv("OLLAMA_TEMP"); tStr != "" {
		if tVal, err := strconv.ParseFloat(tStr, 64); err == nil {
			temperature = tVal
		}
	}
	fmt.Printf("🌡️ Using temperature: %.2f\n", temperature)

	fmt.Printf("🤖 Using model: %s\n", modelName)

	const maxConcurrent = 1 // Max concurrent channels (More Threads = Better Concurrency)
	sem := make(chan struct{}, maxConcurrent)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var evaluations []Evaluation

	for i, jobDesc := range jobDescs {
		wg.Add(1)
		sem <- struct{}{}

		go func(i int, jobDesc string) {
			defer wg.Done()
			defer func() { <-sem }()

			fmt.Printf("🧠 Evaluating job #%d\n", i+1)

			prompt := fmt.Sprintf(`
			I will provide:
			1. My resume.
			2. A job listing.

			Your task is to evaluate my exact fit for the job based strictly on the information provided.

			Requirements:
			- Be extremely detailed and realistic in your scoring.
			- Do NOT inflate the score.
			- Use the full range from 0 to 100.
			- Deduct points for each missing qualification or mismatch.
			- Provide actionable, specific suggestions, not generic tips.
			- Return the result content in an HTML format (eg. <a> for links)

			Format your response EXACTLY as follows:
			---
			Job Title: <title>

			Company: 

			Job Application Link:
			<url>

			Fit Score: <score>/100

			Explanation:
			<why this score was given — be specific and refer to the resume and job listing directly>

			Suggested Resume Changes:
			- <specific change 1>
			- <specific change 2>

			Missing Qualifications:
			- <missing 1>
			- <missing 2>
			---

			Here is my resume:
			===
			%v
			===

			Here is the job listing:
			===
			%v
			===
			`, resumeContent, jobDesc)

			req := Request{
				Model:       modelName,
				Stream:      false,
				Temperature: temperature,
				Messages: []Message{
					{Role: "system", Content: systemInstruction},
					{Role: "user", Content: prompt},
				},
			}

			resp, err := talkToOllama(defaultOllamaURL, req)
			if err != nil {
				fmt.Printf("❌ Error talking to Ollama for job #%d: %v\n", i+1, err)
				return
			}

			cleaned := cleanResponse(resp)
			if !strings.Contains(cleaned, "Fit Score:") {
				fmt.Printf("⚠️ Invalid or malformed response for job #%d, skipping\n", i+1)
			}
			score := extractScore(cleaned)

			formatted := fmt.Sprintf("🔽 Job Evaluation #%d\n%s\n\n", i+1, cleaned)

			mu.Lock()
			evaluations = append(evaluations, Evaluation{
				Score: score,
				Text:  formatted,
			})
			mu.Unlock()
		}(i, jobDesc)
	}

	wg.Wait()

	fmt.Println("📑 Sorting evaluations by score")
	sorted := sortEvaluations(evaluations)

	var outputBuffer strings.Builder
	for _, eval := range sorted {
		outputBuffer.WriteString(eval.Text)
	}

	outputFile = "LinkedinEvaluations.html"
	fmt.Printf("🗑️  Resetting output file: %s\n", outputFile)
	err = os.WriteFile(outputFile, []byte(outputBuffer.String()), 0644)
	if err != nil {
		return err
	}
	fmt.Printf("✅ Text evaluations saved to %s\n", outputFile)

	// Also write HTML file
	htmlFile := "LinkedinEvaluations.html"
	err = writeHTMLFile(htmlFile, sorted)
	if err != nil {
		return err
	}
	fmt.Printf("🌐 HTML evaluations saved to %s\n", htmlFile)

	return nil
}

func extractScore(text string) int {
	fmt.Println("📈 Extracting score from text")
	re := regexp.MustCompile(`(?i)Fit Score:\s*(\d+(?:\.\d+)?)/100`)
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		f, err := strconv.ParseFloat(matches[1], 64)
		if err == nil {
			return int(math.Round(f))
		}
	}
	fmt.Println("⚠️  Score not found or invalid, defaulting to 0")
	return 0
}

func sortEvaluations(evals []Evaluation) []Evaluation {
	fmt.Println("🔃 Sorting evaluations in descending order")
	sort.SliceStable(evals, func(i, j int) bool {
		return evals[i].Score > evals[j].Score
	})
	return evals
}

func talkToOllama(url string, ollamaReq Request) (*Response, error) {
	fmt.Println("📤 Sending request to Ollama")
	reqJSON, err := json.Marshal(&ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	fmt.Printf("📦 Marshalled JSON request size: %d bytes\n", len(reqJSON))

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	fmt.Printf("📥 Received response status: %s\n", res.Status)
	fmt.Printf("📥 Response body size: %d bytes\n", len(bodyBytes))

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s, body: %s", res.Status, string(bodyBytes))
	}

	ollamaResp := Response{}
	err = json.Unmarshal(bodyBytes, &ollamaResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}
	fmt.Println("✅ Successfully unmarshalled response JSON")

	return &ollamaResp, nil
}

func cleanResponse(resp *Response) string {
	fmt.Println("🧹 Cleaning response")
	clean := resp.Message.Content

	// Remove <think> tags
	clean = strings.ReplaceAll(clean, "<think>", "")
	clean = strings.ReplaceAll(clean, "</think>", "")

	// Replace markdown links [text](url) with just url
	linkPattern := regexp.MustCompile(`\[[^\]]*\]\(([^)]+)\)`)
	clean = linkPattern.ReplaceAllString(clean, "$1")

	clean = strings.TrimSpace(clean)
	fmt.Println("🔎 Cleaned response content:")

	return clean
}

func writeHTMLFile(filename string, evaluations []Evaluation) error {
	fmt.Println("🖨️ Generating HTML output")

	var sb strings.Builder
	sb.WriteString("<!DOCTYPE html><html><head><meta charset=\"UTF-8\"><title>Job Evaluations</title>")
	sb.WriteString("<style>body{font-family:sans-serif;padding:20px;} .eval{margin-bottom:40px;padding:20px;border:1px solid #ccc;border-radius:10px;} h2{margin-top:0;} a{color:#0645AD;}</style>")
	sb.WriteString("</head><body><h1>Job Fit Evaluations</h1>")

	for i, eval := range evaluations {
		sb.WriteString("<div class='eval'>")
		sb.WriteString(fmt.Sprintf("<h2>Job Evaluation #%d</h2>", i+1))

		// Convert URLs to clickable links
		htmlContent := convertTextToHTML(eval.Text)
		sb.WriteString(htmlContent)
		sb.WriteString("</div>")
	}

	sb.WriteString("</body></html>")
	return os.WriteFile(filename, []byte(sb.String()), 0644)
}
func convertTextToHTML(text string) string {
	// Escape HTML special chars
	html := strings.ReplaceAll(text, "&", "&amp;")
	html = strings.ReplaceAll(html, "<", "&lt;")
	html = strings.ReplaceAll(html, ">", "&gt;")

	// Replace URLs with <a href="...">
	urlPattern := regexp.MustCompile(`(https?://[^\s<]+)`)
	html = urlPattern.ReplaceAllString(html, `<a href="$1" target="_blank">$1</a>`)

	// Convert newlines to <br>
	html = strings.ReplaceAll(html, "\n", "<br>")

	return html
}

func appendToFile(filename string, content string) error {
	fmt.Printf("📝 Appending to file: %s\n", filename)
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	entry := fmt.Sprintf("\n=============================\n%s\n", content)

	if _, err := f.WriteString(entry); err != nil {
		return err
	}
	fmt.Println("✅ Write successful")
	return nil
}
