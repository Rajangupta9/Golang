package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Configuration for timeouts
type TimeoutConfig struct {
	ShortRequest  time.Duration // For simple requests
	MediumRequest time.Duration // For medium complexity
	LongRequest   time.Duration // For complex/long requests
	HTTPClient    time.Duration // HTTP client timeout
	ServerRead    time.Duration // Server read timeout
	ServerWrite   time.Duration // Server write timeout
}

// Default timeout configuration
var defaultTimeouts = TimeoutConfig{
	ShortRequest:  30 * time.Second,
	MediumRequest: 2 * time.Minute,
	LongRequest:   5 * time.Minute,
	HTTPClient:    6 * time.Minute,
	ServerRead:    30 * time.Second,
	ServerWrite:   6 * time.Minute,
}

type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

type UserPrompt struct {
	Prompt        string `json:"prompt"`
	User          string `json:"user"`
	TimeoutType   string `json:"timeout_type,omitempty"`   // "short", "medium", "long"
	CustomTimeout int    `json:"custom_timeout,omitempty"` // Custom timeout in seconds
}

type ModelResponse struct {
	Response       string `json:"response"`
	Error          string `json:"error,omitempty"`
	ProcessingTime string `json:"processing_time,omitempty"`
	TimeoutUsed    string `json:"timeout_used,omitempty"`
}

type Conversation struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	Prompt    string    `json:"prompt"`
	Response  string    `json:"response"`
	Timestamp time.Time `json:"timestamp"`
}

type UserProfile struct {
	Username      string            `json:"username"`
	Conversations []Conversation    `json:"conversations"`
	PersonalFacts []string          `json:"personal_facts"`
	Preferences   map[string]string `json:"preferences"`
	LastSeen      time.Time         `json:"last_seen"`
}

type MemoryStore struct {
	sync.RWMutex
	users      map[string]*UserProfile
	maxHistory int
	httpClient *http.Client
	timeouts   TimeoutConfig
}

func NewMemoryStore(maxHistory int, timeouts TimeoutConfig) *MemoryStore {
	return &MemoryStore{
		users:      make(map[string]*UserProfile),
		maxHistory: maxHistory,
		timeouts:   timeouts,
		httpClient: &http.Client{
			Timeout: timeouts.HTTPClient,
		},
	}
}

func (ms *MemoryStore) getOrCreateUser(username string) *UserProfile {
	ms.Lock()
	defer ms.Unlock()

	if user, exists := ms.users[username]; exists {
		user.LastSeen = time.Now()
		return user
	}

	user := &UserProfile{
		Username:      username,
		Conversations: make([]Conversation, 0),
		PersonalFacts: make([]string, 0),
		Preferences:   make(map[string]string),
		LastSeen:      time.Now(),
	}
	ms.users[username] = user
	return user
}

func (ms *MemoryStore) addConversation(username, prompt, response string) {
	ms.Lock()
	defer ms.Unlock()

	user := ms.users[username]
	if user == nil {
		return
	}

	conversation := Conversation{
		ID:        fmt.Sprintf("%s_%d", username, time.Now().Unix()),
		User:      username,
		Prompt:    prompt,
		Response:  response,
		Timestamp: time.Now(),
	}

	user.Conversations = append(user.Conversations, conversation)

	if len(user.Conversations) > ms.maxHistory {
		user.Conversations = user.Conversations[len(user.Conversations)-ms.maxHistory:]
	}

	ms.extractPersonalInfo(user, prompt, response)
}

func (ms *MemoryStore) extractPersonalInfo(user *UserProfile, prompt, _ string) {
	personalKeywords := []string{"my name is", "i am", "i like", "i work", "i live", "my job", "my hobby"}

	lowerPrompt := strings.ToLower(prompt)
	for _, keyword := range personalKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			fact := strings.TrimSpace(prompt)
			if !contains(user.PersonalFacts, fact) {
				user.PersonalFacts = append(user.PersonalFacts, fact)
			}
		}
	}

	if len(user.PersonalFacts) > 20 {
		user.PersonalFacts = user.PersonalFacts[len(user.PersonalFacts)-20:]
	}
}

// Enhanced context building with size limits based on timeout type
func (ms *MemoryStore) buildContext(username, currentPrompt, timeoutType string) string {
	ms.RLock()
	defer ms.RUnlock()

	user := ms.users[username]
	if user == nil {
		return currentPrompt
	}

	var contextBuilder strings.Builder
	var maxHistoryItems int

	// Adjust context size based on timeout type
	switch timeoutType {
	case "short":
		maxHistoryItems = 2
	case "medium":
		maxHistoryItems = 5
	case "long":
		maxHistoryItems = 10
	default:
		maxHistoryItems = 3
	}

	if len(user.PersonalFacts) > 0 {
		contextBuilder.WriteString("Personal Information about " + username + ":\n")
		for _, fact := range user.PersonalFacts {
			contextBuilder.WriteString("- " + fact + "\n")
		}
		contextBuilder.WriteString("\n")
	}

	recentConversations := user.Conversations
	if len(recentConversations) > maxHistoryItems {
		recentConversations = recentConversations[len(recentConversations)-maxHistoryItems:]
	}

	if len(recentConversations) > 0 {
		contextBuilder.WriteString("Recent Conversation History:\n")
		for _, conv := range recentConversations {
			contextBuilder.WriteString(fmt.Sprintf("User: %s\nAssistant: %s\n\n", conv.Prompt, conv.Response))
		}
	}

	contextBuilder.WriteString("Current Question: " + currentPrompt)

	// Limit total context size
	context := contextBuilder.String()
	maxContextSize := getMaxContextSize(timeoutType)
	if len(context) > maxContextSize {
		context = context[len(context)-maxContextSize:]
	}

	return context
}

func getMaxContextSize(timeoutType string) int {
	switch timeoutType {
	case "short":
		return 2000
	case "medium":
		return 6000
	case "long":
		return 15000
	default:
		return 4000
	}
}

func (ms *MemoryStore) getUserStats(username string) map[string]interface{} {
	ms.RLock()
	defer ms.RUnlock()

	user := ms.users[username]
	if user == nil {
		return nil
	}

	return map[string]interface{}{
		"username":            user.Username,
		"total_conversations": len(user.Conversations),
		"personal_facts":      len(user.PersonalFacts),
		"last_seen":           user.LastSeen,
		"preferences":         user.Preferences,
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Determine timeout based on request characteristics
func determineTimeout(userInput UserPrompt, timeouts TimeoutConfig) (time.Duration, string) {
	// Use custom timeout if provided
	if userInput.CustomTimeout > 0 {
		customDuration := time.Duration(userInput.CustomTimeout) * time.Second
		// Cap custom timeout at 10 minutes for safety
		if customDuration > 10*time.Minute {
			customDuration = 10 * time.Minute
		}
		return customDuration, fmt.Sprintf("custom_%ds", userInput.CustomTimeout)
	}

	// Use specified timeout type
	if userInput.TimeoutType != "" {
		switch userInput.TimeoutType {
		case "short":
			return timeouts.ShortRequest, "short_30s"
		case "medium":
			return timeouts.MediumRequest, "medium_2m"
		case "long":
			return timeouts.LongRequest, "long_5m"
		}
	}

	// Auto-determine based on prompt characteristics
	promptLen := len(userInput.Prompt)

	// Check for complexity indicators
	complexityKeywords := []string{
		"analyze", "explain in detail", "write a story", "create a plan",
		"summarize", "compare", "research", "detailed analysis",
		"step by step", "comprehensive", "thorough", "elaborate", "create",
	}

	isComplex := false
	lowerPrompt := strings.ToLower(userInput.Prompt)
	for _, keyword := range complexityKeywords {
		if strings.Contains(lowerPrompt, keyword) {
			isComplex = true
			break
		}
	}

	// Determine timeout based on length and complexity
	if promptLen > 1000 || isComplex {
		return timeouts.LongRequest, "auto_long_5m"
	} else if promptLen > 300 {
		return timeouts.MediumRequest, "auto_medium_2m"
	}

	return timeouts.ShortRequest, "auto_short_30s"
}

var memoryStore = NewMemoryStore(50, defaultTimeouts)

func queryLLaMA(ctx context.Context, fullPrompt string) (string, error) {
	url := "http://localhost:11434/api/generate"

	requestBody, err := json.Marshal(OllamaRequest{
		Model:  "llama3",
		Prompt: fullPrompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("request timeout before sending")
	default:
	}

	resp, err := memoryStore.httpClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("request timeout: LLaMA took too long to respond")
		}
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLaMA API returned status %d", resp.StatusCode)
	}

	
	bodyReader := io.LimitReader(resp.Body, 5*1024*1024) // 5MB limit
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result OllamaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return result.Response, nil
}

func handlePrompt(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Only POST method allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var userInput UserPrompt
	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		json.NewEncoder(w).Encode(ModelResponse{Error: "Invalid JSON format"})
		return
	}

	if userInput.Prompt == "" || userInput.User == "" {
		json.NewEncoder(w).Encode(ModelResponse{Error: "Prompt and user fields are required"})
		return
	}

	// Validate prompt length (increased for long requests)
	maxPromptLength := 10000 // 10KB for long requests
	if len(userInput.Prompt) > maxPromptLength {
		json.NewEncoder(w).Encode(ModelResponse{
			Error: fmt.Sprintf("Prompt too long (max %d characters)", maxPromptLength),
		})
		return
	}

	memoryStore.getOrCreateUser(userInput.User)

	// Determine appropriate timeout
	requestTimeout, timeoutLabel := determineTimeout(userInput, memoryStore.timeouts)

	fullPrompt := memoryStore.buildContext(userInput.User, userInput.Prompt, userInput.TimeoutType)

	ctx, cancel := context.WithTimeout(r.Context(), requestTimeout)
	defer cancel()

	log.Printf("Processing request for user %s with timeout %s", userInput.User, timeoutLabel)

	type result struct {
		response string
		err      error
	}

	resultChan := make(chan result, 1)

	go func() {
		response, err := queryLLaMA(ctx, fullPrompt)
		resultChan <- result{response: response, err: err}
	}()

	select {
	case res := <-resultChan:
		processingTime := time.Since(startTime)

		if res.err != nil {
			log.Printf("Error querying LLaMA for user %s: %v", userInput.User, res.err)

			if strings.Contains(res.err.Error(), "timeout") {
				json.NewEncoder(w).Encode(ModelResponse{
					Error:          "Request timeout - try using 'timeout_type': 'long' or 'custom_timeout': 300 for complex requests",
					ProcessingTime: processingTime.String(),
					TimeoutUsed:    timeoutLabel,
				})
			} else {
				json.NewEncoder(w).Encode(ModelResponse{
					Error:          "AI service temporarily unavailable",
					ProcessingTime: processingTime.String(),
					TimeoutUsed:    timeoutLabel,
				})
			}
			return
		}

		memoryStore.addConversation(userInput.User, userInput.Prompt, res.response)
		log.Printf("User %s: %s (processed in %s)", userInput.User, userInput.Prompt, processingTime)

		json.NewEncoder(w).Encode(ModelResponse{
			Response:       res.response,
			ProcessingTime: processingTime.String(),
			TimeoutUsed:    timeoutLabel,
		})

	case <-ctx.Done():
		processingTime := time.Since(startTime)
		log.Printf("Request timeout for user %s after %s", userInput.User, processingTime)
		json.NewEncoder(w).Encode(ModelResponse{
			Error:          "Request timeout - try using 'timeout_type': 'long' or increase 'custom_timeout' for complex requests",
			ProcessingTime: processingTime.String(),
			TimeoutUsed:    timeoutLabel,
		})
		return
	}
}

func handleUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Only GET method allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	username := r.URL.Query().Get("user")
	if username == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "User parameter is required"})
		return
	}

	stats := memoryStore.getUserStats(username)
	if stats == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "User not found"})
		return
	}

	json.NewEncoder(w).Encode(stats)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"users":     len(memoryStore.users),
		"timeouts": map[string]string{
			"short_requests":  memoryStore.timeouts.ShortRequest.String(),
			"medium_requests": memoryStore.timeouts.MediumRequest.String(),
			"long_requests":   memoryStore.timeouts.LongRequest.String(),
			"http_client":     memoryStore.timeouts.HTTPClient.String(),
		},
	})
}

func handleTimeoutInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	info := map[string]interface{}{
		"timeout_types": map[string]string{
			"short":  "30 seconds - for simple questions",
			"medium": "2 minutes - for moderate complexity",
			"long":   "5 minutes - for complex analysis/generation",
		},
		"custom_timeout": "Use 'custom_timeout' field with seconds (max 600)",
		"auto_detection": "System auto-detects based on prompt length and complexity",
		"example_requests": map[string]interface{}{
			"short_request": map[string]string{
				"prompt":       "Hello, how are you?",
				"user":         "username",
				"timeout_type": "short",
			},
			"long_request": map[string]interface{}{
				"prompt":         "Write a detailed analysis of...",
				"user":           "username",
				"timeout_type":   "long",
				"custom_timeout": 300,
			},
		},
	}

	json.NewEncoder(w).Encode(info)
}

func main() {
	// Allow configuration via environment variables
	if customShort := getEnvDuration("SHORT_TIMEOUT", defaultTimeouts.ShortRequest); customShort != 0 {
		defaultTimeouts.ShortRequest = customShort
	}
	if customMedium := getEnvDuration("MEDIUM_TIMEOUT", defaultTimeouts.MediumRequest); customMedium != 0 {
		defaultTimeouts.MediumRequest = customMedium
	}
	if customLong := getEnvDuration("LONG_TIMEOUT", defaultTimeouts.LongRequest); customLong != 0 {
		defaultTimeouts.LongRequest = customLong
	}

	// Recreate memory store with updated timeouts
	memoryStore = NewMemoryStore(50, defaultTimeouts)

	http.HandleFunc("/prompt", handlePrompt)
	http.HandleFunc("/profile", handleUserProfile)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/timeout-info", handleTimeoutInfo)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  defaultTimeouts.ServerRead,
		WriteTimeout: defaultTimeouts.ServerWrite,
		IdleTimeout:  2 * time.Minute,
	}

	log.Fatal(server.ListenAndServe())
}

// Helper function to get duration from environment variable
func getEnvDuration(envVar string, defaultValue time.Duration) time.Duration {
	
	return 0
}
