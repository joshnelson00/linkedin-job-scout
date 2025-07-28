package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
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

func getJobEvaluation(jobDesc string) string {
	fmt.Println("🔍 Starting getJobEvaluation")

	resumeBytes, err := os.ReadFile("resume.txt")
	if err != nil {
		fmt.Printf("❌ Error reading resume.txt: %v\n", err)
		return ""
	}
	fmt.Println("✅ Successfully read resume.txt")

	resumeContent := string(resumeBytes)
	fmt.Printf("📝 Resume content length: %d characters\n", len(resumeContent))
	fmt.Printf("📄 Job description length: %d characters\n", len(jobDesc))

	prompt := fmt.Sprintf(`
	You are an expert career advisor and resume evaluator.

	I will provide:
	1. My resume.
	2. A job listing.

	Your task is to evaluate my fit for the job and return a response in the following EXACT format:

	---
	Job Title: <title>

	Job Application Link: <url>

	Fit Score: <score>/10

	Explanation:
	<why this score was given>

	Suggested Resume Changes:
	- <change 1>
	- <change 2>
	- <etc.>

	Missing Qualifications:
	- <missing qualification 1>
	- <missing qualification 2>
	- <etc.>

	Optional Cover Letter Opening:
	"<suggested opening>"
	---

	Here is my resume:
	===
	%v
	===

	Here is the job listing:
	===
	%v
	===`, resumeContent, jobDesc)

	msg := Message{
		Role:    "user",
		Content: prompt,
	}

	modelName := os.Getenv("OLLAMA_MODEL")
	if modelName == "" {
		modelName = "llama3.2" // 🧠 Use a lighter model instead of "deepseek-r1"
	}

	req := Request{
		Model:    modelName,
		Stream:   false,
		Messages: []Message{msg},
	}

	fmt.Println("📡 Sending request to Ollama API...")
	resp, err := talkToOllama(defaultOllamaURL, req)
	if err != nil {
		fmt.Printf("❌ Error talking to Ollama: %v\n", err)
		return ""
	}
	fmt.Println("✅ Received response from Ollama")

	// 💤 Throttle here to reduce CPU load (e.g., between jobs)
	fmt.Println("⏳ Sleeping for 3 seconds to reduce load...")
	time.Sleep(3 * time.Second)

	cleanedResponse := cleanResponse(resp)
	fmt.Printf("🧹 Cleaned response length: %d characters\n", len(cleanedResponse))

	// 📝 Save evaluation to file
	outputFile := "evaluations.txt"
	err = appendToFile(outputFile, cleanedResponse)
	if err != nil {
		fmt.Printf("❌ Failed to write to %s: %v\n", outputFile, err)
	} else {
		fmt.Printf("💾 Saved evaluation to %s\n", outputFile)
	}

	return cleanedResponse
}


func talkToOllama(url string, ollamaReq Request) (*Response, error) {
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
	clean := resp.Message.Content
	fmt.Println("🔎 Original response content:")
	fmt.Println(clean)

	clean = strings.ReplaceAll(clean, "<think>", "")
	clean = strings.ReplaceAll(clean, "</think>", "")
	clean = strings.TrimSpace(clean)

	fmt.Println("🔎 Cleaned response content:")
	fmt.Println(clean)

	return clean
}

func appendToFile(filename string, content string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Add a clear separator between entries
	entry := fmt.Sprintf("\n=============================\n%s\n", content)

	if _, err := f.WriteString(entry); err != nil {
		return err
	}
	return nil
}
