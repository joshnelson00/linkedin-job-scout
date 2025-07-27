package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func main() {
	fmt.Println("Creating message...")
	msg := Message{
		Role:    "user",
		Content: "Why is the sky blue?",
	}

	fmt.Println("Constructing request...")
	req := Request{
		Model:    "deepseek-r1", // Ensure this model is installed and supports /api/chat
		Stream:   false,
		Messages: []Message{msg},
	}

	fmt.Println("Sending request to Ollama...")
	resp, err := talkToOllama(defaultOllamaURL, req)
	if err != nil {
		fmt.Printf("❌ Error talking to Ollama: %v\n", err)
		return
	}

	fmt.Println("Cleaning response...")
	cleanedReponse := cleanResponse(resp)
	fmt.Printf("✅ Response: %s\n", cleanedReponse)
}

func talkToOllama(url string, ollamaReq Request) (*Response, error) {
	fmt.Println("Marshaling request to JSON...")
	reqJSON, err := json.Marshal(&ollamaReq)
	if err != nil {
		fmt.Println("❌ Failed to marshal request")
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	fmt.Println("Creating HTTP request...")
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqJSON))
	if err != nil {
		fmt.Println("❌ Failed to create HTTP request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := http.Client{}
	fmt.Printf("POSTing to %s...\n", url)
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("❌ Failed to send HTTP request")
		return nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer res.Body.Close()

	fmt.Println("Reading response body...")
	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("❌ Failed to read response body")
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		fmt.Printf("❌ Unexpected status code: %s\n", res.Status)
		fmt.Printf("Response body: %s\n", string(bodyBytes))
		return nil, fmt.Errorf("unexpected status: %s, body: %s", res.Status, string(bodyBytes))
	}

	fmt.Println("Unmarshaling JSON response...")
	ollamaResp := Response{}
	err = json.Unmarshal(bodyBytes, &ollamaResp)
	if err != nil {
		fmt.Printf("❌ Failed to decode response JSON: %v\n", err)
		fmt.Printf("Raw body: %s\n", string(bodyBytes))
		return nil, fmt.Errorf("failed to decode response JSON: %w", err)
	}

	fmt.Println("✅ Response unmarshaled successfully")
	return &ollamaResp, nil
}

func cleanResponse(resp *Response) string {
	clean := resp.Message.Content
	clean = strings.ReplaceAll(clean, "<think>", "")
	clean = strings.ReplaceAll(clean, "</think>", "")
	clean = strings.TrimSpace(clean)

	return clean
}
