package main

import (
	"encoding/json"
	"fmt"
	"html"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp  string              `json:"timestamp"`
	IP         string              `json:"ip"`
	Method     string              `json:"method"`
	Path       string              `json:"path"`
	UserAgent  string              `json:"user_agent"`
	Headers    map[string][]string `json:"headers,omitempty"`
	Event      string              `json:"event"`
	StatusCode int                 `json:"status_code"`
}

var (
	accessLog     *log.Logger
	errorLog      *log.Logger

	apacheVersions = []string{
		"Apache/2.2.16 (Unix)",
		"Apache/2.4.1 (Win32)",
		"Apache/2.0.63 (Unix)",
	}

	phpVersions = []string{
		"PHP/5.3.28",
		"PHP/5.2.17",
		"PHP/4.4.9",
	}

	loginResponses = []string{
		"Login successful. Redirecting...",
		"Incorrect username.",
		"Incorrect password.",
		"Account temporarily locked. Please try again later.",
	}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func initLoggers() {
	accessFile, err := os.OpenFile("honeypot_access.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf(`{"level":"fatal","message":"cannot open access log file","error":"%v"}`, err)
	}

	errorFile, err := os.OpenFile("honeypot_errors.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf(`{"level":"fatal","message":"cannot open error log file","error":"%v"}`, err)
	}

	accessLog = log.New(accessFile, "", 0)
	errorLog = log.New(errorFile, "", 0)
}

func addRandomDelay() {
	time.Sleep(time.Duration(1000+rand.Intn(1000)) * time.Millisecond)
}

func addFakeHeaders(w http.ResponseWriter) {
	w.Header().Set("Server", apacheVersions[rand.Intn(len(apacheVersions))])
	w.Header().Set("X-Powered-By", phpVersions[rand.Intn(len(phpVersions))])
}

func sanitizeInput(input string) string {
	return html.EscapeString(input)
}

func sanitizeHeaders(headers map[string][]string) map[string][]string {
	sanitizedHeaders := make(map[string][]string)
	for key, values := range headers {
		sanitizedKey := sanitizeInput(key)
		var sanitizedValues []string
		for _, value := range values {
			sanitizedValues = append(sanitizedValues, sanitizeInput(value))
		}
		sanitizedHeaders[sanitizedKey] = sanitizedValues
	}
	return sanitizedHeaders
}

func logJSON(entry LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		errorLog.Printf(`{"level":"error","message":"failed to marshal log entry","error":"%v"}`, err)
		return
	}
	accessLog.Println(string(data))
}

func logRequest(r *http.Request, statusCode int, event string) {
	entry := LogEntry{
		Timestamp:  time.Now().Format(time.RFC3339),
		IP:         sanitizeInput(r.RemoteAddr),
		Method:     sanitizeInput(r.Method),
		Path:       sanitizeInput(r.URL.Path),
		UserAgent:  sanitizeInput(r.UserAgent()),
		Headers:    sanitizeHeaders(r.Header),
		Event:      event,
		StatusCode: statusCode,
	}

	logJSON(entry)
}

func handler(w http.ResponseWriter, r *http.Request) {
	addRandomDelay()
	addFakeHeaders(w)

	if r.URL.Path == "/" || r.URL.Path == "/index.html" || r.URL.Path == "/index.php" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head>
	<title>Welcome to MyBank Online</title>
	<style>
		body { font-family: Arial, sans-serif; background-color: #e0e0e0; color: #000; }
		.container { width: 760px; margin: 40px auto; background: white; padding: 20px; border: 1px solid #aaa; }
		.header { background-color: #003366; color: white; padding: 10px; text-align: center; }
		.nav { background-color: #d0d0d0; padding: 8px; font-size: 14px; }
		.footer { font-size: 12px; text-align: center; margin-top: 20px; color: #555; }
	</style>
</head>
<body>
	<div class="container">
		<div class="header">
			<h1>MyBank Online Services</h1>
			<p>Secure Internet Banking</p>
		</div>
		<div class="nav">
			<a href="/login.php">Login</a> |
			<a href="/index.html">Home</a> |
			<a href="/contact.php">Contact Us</a>
		</div>
		<p>Welcome to MyBank's secure online banking portal. Please <a href="/login.php">log in</a> to access your accounts.</p>
		<p>Online banking lets you check balances, view transactions, transfer funds, and pay bills â€” all from the comfort of your home.</p>
		<hr>
		<p><strong>Notice:</strong> This service requires Internet Explorer 6.0 or higher with 128-bit encryption enabled.</p>
		<div class="footer">
			&copy; 2003 MyBank Corp. All Rights Reserved.<br>
		</div>
	</div>
</body>
</html>
		`)
		logRequest(r, http.StatusOK, "served index page")
		return
	}

	if r.URL.Path == "/robots.txt" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "User-agent: *\nDisallow: /admin.php\nDisallow: /cgi-bin/\nDisallow: /secure/\n")
		logRequest(r, http.StatusOK, "robots.txt served")
		return
	}

	if strings.Contains(r.URL.Path, "cgi-bin") {
		http.Error(w, "500 Internal Server Error: Something went wrong", http.StatusInternalServerError)
		logRequest(r, http.StatusInternalServerError, "cgi-bin access attempt")
		return
	}

	if r.URL.Path == "/admin.php" {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `<!DOCTYPE html><html><head><title>Admin Dashboard</title></head><body>
<h3>Current Database Entires: 4214</h3>
<h2>Admin Control Panel</h2><ul>
  <li><a href="/login.php">View All Accounts</a></li>
  <li><a href="/login.php">Modify Balances</a></li>
  <li><a href="/login.php">Download Logs</a></li>
  <li><a href="/login.php">System Settings</a></li>
</ul><p>Loading secure panel...</p></body></html>`)
		logRequest(r, http.StatusOK, "admin.php accessed - loading screen")
		go func(ip, ua string) {
			time.Sleep(10 * time.Second)
			entry := LogEntry{Timestamp: time.Now().Format(time.RFC3339), IP: r.RemoteAddr, Method: r.Method, Path: r.URL.Path, UserAgent: ua, Event: "admin.php delayed response: access denied", StatusCode: 403}
			logJSON(entry)
		}(r.RemoteAddr, r.UserAgent())
		return
	}

	if r.URL.Path == "/login.php" {
		switch r.Method {
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><title>Login</title></head><body>
<h2>Login</h2><form method="POST" action="/login.php">
  <label for="username">Username:</label><br>
  <input type="text" id="username" name="username"><br><br>
  <label for="password">Password:</label><br>
  <input type="password" id="password" name="password"><br><br>
  <input type="submit" value="Login">
</form></body></html>`)
			logRequest(r, http.StatusOK, "login page served")
			return
		case http.MethodPost:
			username := sanitizeInput(r.FormValue("username"))
			password := sanitizeInput(r.FormValue("password"))
			response := loginResponses[rand.Intn(len(loginResponses))]
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, response)
			logRequest(r, http.StatusOK, fmt.Sprintf("login attempt: %s (username: %s, password: %s)", response, username, password))
			return
		}
	}
}

func main() {
	initLoggers()
	http.HandleFunc("/", handler)
	http.HandleFunc("/index.html", handler)
	http.HandleFunc("/index.php", handler)
	http.HandleFunc("/login.php", handler)
	http.HandleFunc("/robots.txt", handler)
	http.HandleFunc("/admin.php", handler)

	log.Println("Honeypot is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

