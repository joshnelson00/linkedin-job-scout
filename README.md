# LinkedIn Job Scout

An AI-powered job discovery and ranking platform built in Go that automates the process of finding, analyzing, and prioritizing software engineering opportunities.

LinkedIn job searching is often a repetitive workflow: searching hundreds of listings, comparing requirements against your experience, and manually determining which positions are worth applying to.

LinkedIn Job Scout automates this process by collecting job postings, caching results, analyzing compatibility with a candidate profile using local LLM inference, and generating ranked job recommendations.

## Features

- 🔎 Automated job discovery through third-party job APIs
- ⚡ Concurrent job processing using Go goroutines
- 🧠 LLM-powered resume/job compatibility analysis
- 📊 Automated job fit scoring (0-100)
- 💡 AI-generated improvement suggestions for resume alignment
- 🚀 Redis-backed caching to reduce duplicate processing
- 📧 Automated email delivery of ranked job recommendations
- 🔄 End-to-end pipeline from job retrieval → analysis → notification

### Concurrent Processing Pipeline

Built a concurrent backend pipeline using Go goroutines and channels to process multiple job postings efficiently while controlling API usage.

Key concepts:
- Worker-style processing
- Rate limiting
- Asynchronous job evaluation
- Fault isolation

### Intelligent Job Matching

Each job posting is analyzed against a candidate resume to determine:

- Technical skill alignment
- Experience relevance
- Missing qualifications
- Suggested resume improvements

The system produces a compatibility score and actionable recommendations.

### Redis Caching Layer

Implemented Redis caching to:

- Prevent duplicate job analysis
- Reduce unnecessary API calls
- Improve processing speed
- Maintain state between executions

## Example Workflow

1. Retrieve new job postings from external APIs
2. Normalize job descriptions
3. Check Redis cache for existing analysis
4. Send new postings to the LLM pipeline
5. Generate compatibility scores and recommendations
6. Rank opportunities
7. Deliver results through email


This project was built to explore practical applications of:

- Distributed backend systems
- Concurrent programming in Go
- AI-assisted decision making
- API integration
- Infrastructure automation

Rather than replacing the job search process, LinkedIn Job Scout acts as an intelligent filtering layer that helps candidates spend more time applying to high-value opportunities.

## Running Locally

### Prerequisites

Install the following:

- Go 1.22+
- Redis
- Ollama

Verify installations:

```bash
go version
redis-server --version
ollama --version
```

### Set up .env
```
SCRAPINGDOG_API_KEY=<your_api_key>

REDIS_HOST=localhost
REDIS_PORT=6379

OLLAMA_MODEL=llama3.2
```
## Running Locally

### Start Dependencies

Start Redis:

```bash
redis-server
```

Start Ollama and Download the Model:

```bash
ollama serve
ollama pull llama3.2
```

Pull Go Libraries:

```bash
go mod downlaod
```
### Run Application

```bash
go run .
```
