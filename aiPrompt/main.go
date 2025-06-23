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

// Request/Response structures
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

type UserPrompt struct {
	Prompt string `json:"prompt"`
	User   string `json:"user"`
}

type ModelResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

// Enhanced memory structures
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

// Thread-safe memory store
type MemoryStore struct {
	sync.RWMutex
	users      map[string]*UserProfile
	maxHistory int
	httpClient *http.Client
}

func NewMemoryStore(maxHistory int) *MemoryStore {
	return &MemoryStore{
		users:      make(map[string]*UserProfile),
		maxHistory: maxHistory,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get or create user profile
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

// Add conversation to user's history
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

	// Maintain conversation history limit
	if len(user.Conversations) > ms.maxHistory {
		user.Conversations = user.Conversations[len(user.Conversations)-ms.maxHistory:]
	}

	// Extract and store personal information
	ms.extractPersonalInfo(user, prompt, response)
}

// Extract personal information from conversations
func (ms *MemoryStore) extractPersonalInfo(user *UserProfile, prompt, _ string) {
	// Simple keyword-based extraction (can be enhanced with NLP)
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

// Build context for LLM
func (ms *MemoryStore) buildContext(username, currentPrompt string) string {
	ms.RLock()
	defer ms.RUnlock()

	user := ms.users[username]
	if user == nil {
		return currentPrompt
	}

	var contextBuilder strings.Builder

	if len(user.PersonalFacts) > 0 {
		contextBuilder.WriteString("Personal Information about " + username + ":\n")
		for _, fact := range user.PersonalFacts {
			contextBuilder.WriteString("- " + fact + "\n")
		}
		contextBuilder.WriteString("\n")
	}

	recentConversations := user.Conversations
	if len(recentConversations) > 5 {
		recentConversations = recentConversations[len(recentConversations)-5:]
	}

	if len(recentConversations) > 0 {
		contextBuilder.WriteString("Recent Conversation History:\n")
		for _, conv := range recentConversations {
			contextBuilder.WriteString(fmt.Sprintf("User: %s\nAssistant: %s\n\n", conv.Prompt, conv.Response))
		}
	}

	contextBuilder.WriteString("Current Question: " + currentPrompt)

	return contextBuilder.String()
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

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

var memoryStore = NewMemoryStore(50)

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

	resp, err := memoryStore.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("LLaMA API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
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
	w.Header().Set("Content-Type", "application/json")

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

	memoryStore.getOrCreateUser(userInput.User)

	fullPrompt := memoryStore.buildContext(userInput.User, userInput.Prompt)

	ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
	defer cancel()

	response, err := queryLLaMA(ctx, fullPrompt)
	if err != nil {
		log.Printf("Error querying LLaMA for user %s: %v", userInput.User, err)
		json.NewEncoder(w).Encode(ModelResponse{Error: "Failed to get response from AI model"})
		return
	}

	memoryStore.addConversation(userInput.User, userInput.Prompt, response)

	log.Printf("User %s: %s", userInput.User, userInput.Prompt)
	log.Printf("Response: %s", response)

	json.NewEncoder(w).Encode(ModelResponse{Response: response})
}

func handleUserProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

func main() {
	http.HandleFunc("/prompt", handlePrompt)
	http.HandleFunc("/profile", handleUserProfile)

	server := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}
