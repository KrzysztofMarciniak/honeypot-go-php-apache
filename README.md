# 🕵️ Go Honeypot

This project is a simple **honeypot web server** written in Go, designed to simulate a fake online banking site. It logs access attempts, suspicious activity, and fake login interactions in structured JSON format for analysis.

## 🧠 Features

- Simulates an old banking site (`MyBank Online`) with login and admin endpoints.
- Logs all HTTP interactions in JSON (`honeypot_access.json`, `honeypot_errors.json`).
```json
{
	"timestamp":"2025-05-04T09:19:08Z",
	"ip":"172.21.0.1:53090",
	"method":"GET",
	"path":"/",
	"user_agent":"curl/8.11.1",
	"headers":{"Accept":["*/*"],
	"User-Agent":["curl/8.11.1"]},
	"event":"served index page",
	"status_code":200
}
```
- Mimics a vulnerable system with fake `admin.php`, `cgi-bin`, and outdated headers (`Apache`, `PHP`).
- Random response delays and server headers to increase authenticity.
- Dockerized for easy deployment and resource limits.

## 🚀 Quick Start

### 🔧 Requirements

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/)

### 🐳 Run via Docker Compose
```bash
docker compose up
```
Once running, access the honeypot at:
```bash
http://localhost:8080
```
If it does not immediately work, wait at least a minute.

### 🗂️ Project Structure
```bash
├── honeypot.go              # Main Go server file
├── compose.yml              # Docker Compose setup
├── honeypot_access.json     # Logs all HTTP requests
├── honeypot_errors.json     # Logs application errors
```

📝 Notes
- This honeypot is not intended for production. It's a research/demo tool.
- Ensure it's run in an isolated environment (use the provided bridge network).
- Avoid exposing it to the public internet unless secured.


