# LinkedIn Job Scout AI Evaluator

A Go-based tool that automatically fetches job listings and uses an LLM (via [Ollama](https://ollama.com)) to evaluate which roles best match your resume. Ideal for job seekers and AI-enhanced career tools.

---

## 🚀 Features

- 🔍 **Fetches job listings** from an API (e.g., ScrapingDog)
- 📄 **Loads and parses your resume** from `resume.html`
- 🤖 **Uses a local LLM** (via Ollama) to evaluate job fit
- ⚖️ **Scores and sorts jobs** based on AI evaluation
- 📁 Outputs a ranked list to `LinkedinEvaluations.txt`
- 💾 **Caching** for job descriptions to avoid repeat work and save on API cost
- 📊 Designed with concurrency and token efficiency in mind

---


