package main // Declares the package name as main, which is required for an executable program

import (
	"bytes"        // Imports the bytes package for buffer manipulation
	"encoding/json" // Imports encoding/json for JSON encoding and decoding
	"fmt"           // Imports fmt for formatted I/O
	"net/http"      // Imports net/http for HTTP server and client implementations
)

// Request and Message struct for sending to Ollama
type Request struct {
	Model    string    `json:"model"`    // Model name to use in the request
	Messages []Message `json:"messages"` // List of messages for the chat
	Stream   bool      `json:"stream"`   // Whether to stream the response
}

type Message struct {
	Role    string `json:"role"`    // Role of the message sender (e.g., "user")
	Content string `json:"content"` // Content of the message
}

// Response struct for receiving from Ollama
type Response struct {
	Model      string `json:"model"`        // Model name used in the response
	Created_at string `json:"created_at"`   // Timestamp of response creation
	Message    struct {
		Role    string `json:"role"`    // Role of the message sender in the response
		Content string `json:"content"` // Content of the generated message
	} `json:"message"`
	Done bool `json:"done"` // Indicates if the response is complete
}

// chatHandler handles HTTP POST requests for chat interactions with the Ollama API.
// 
// This handler expects a JSON payload from the client with the following structure:
//   {
//     "prompt": "<user input prompt>"
//   }
//
// The handler performs the following steps:
//   1. Validates that the request method is POST. If not, responds with HTTP 405 Method Not Allowed.
//   2. Decodes the JSON request body to extract the user's prompt. If decoding fails, responds with HTTP 400 Bad Request.
//   3. Constructs a request payload for the Ollama API, specifying the model ("llama3.2") and including the user's prompt as a message.
//   4. Marshals the request payload to JSON. If marshalling fails, responds with HTTP 500 Internal Server Error.
//   5. Sends the JSON payload to the Ollama API at http://localhost:11434/api/chat using an HTTP POST request.
//      If the request fails, responds with HTTP 500 Internal Server Error.
//   6. Decodes the JSON response from the Ollama API. If decoding fails, responds with HTTP 500 Internal Server Error.
//   7. Extracts the generated message from the Ollama response and sends it back to the client as a JSON object:
//      {
//        "response": "<Ollama generated response>"
//      }
//
// The handler sets the "Content-Type" header to "application/json" for the response.
//
// Expected types:
//   - Request: Struct representing the payload sent to Ollama, including model, messages, and stream flag.
//   - Message: Struct representing a single message with role and content.
//   - Response: Struct representing the response from Ollama, including the generated message.
//
// Errors are logged and returned to the client with appropriate HTTP status codes.
func chatHandler(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests
	if r.Method != http.MethodPost { // Checks if the HTTP method is POST
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed) // Responds with 405 if not POST
		return // Exits the handler
	}

	// Read request body sent to our server
	var clientRequest struct {
		Prompt string `json:"prompt"` // Struct to hold the incoming prompt
	}

	err := json.NewDecoder(r.Body).Decode(&clientRequest) // Decodes the JSON request body into clientRequest
	if err != nil { // Checks for decoding errors
		http.Error(w, "Invalid request body", http.StatusBadRequest) // Responds with 400 if decoding fails
		return // Exits the handler
	}

	// Build request to Ollama
	ollamaRequest := Request{
		Model: "llama3.2", // Sets the model to "llama3.2"
		Messages: []Message{
			{
				Role:    "user", // Sets the role as "user"
				Content: clientRequest.Prompt, // Sets the content to the user's prompt
			},
		},
		Stream: false, // Disables streaming
	}

	jsonData, err := json.Marshal(ollamaRequest) // Marshals the request struct to JSON
	if err != nil { // Checks for marshalling errors
		http.Error(w, "Error marshalling JSON", http.StatusInternalServerError) // Responds with 500 if marshalling fails
		return // Exits the handler
	}

	// Send request to Ollama
	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(jsonData)) // Sends the JSON to Ollama API
	if err != nil { // Checks for HTTP request errors
		http.Error(w, "Error making request to Ollama", http.StatusInternalServerError) // Responds with 500 if request fails
		fmt.Println("Error making request:", err) // Logs the error
		return // Exits the handler
	}
	defer resp.Body.Close() // Ensures the response body is closed

	// Read Ollama response
	var ollamaResponse Response // Struct to hold the Ollama response
	err = json.NewDecoder(resp.Body).Decode(&ollamaResponse) // Decodes the JSON response into ollamaResponse
	if err != nil { // Checks for decoding errors
		http.Error(w, "Error decoding Ollama response", http.StatusInternalServerError) // Responds with 500 if decoding fails
		return // Exits the handler
	}

	// Send response back to client
	responseToClient := map[string]string{
		"response": ollamaResponse.Message.Content, // Prepares the response with the generated content
	}

	w.Header().Set("Content-Type", "application/json") // Sets the response content type to JSON
	json.NewEncoder(w).Encode(responseToClient) // Encodes and sends the response to the client
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintln(w, `<h1>Ollama Chat API Demo</h1>
	<p>POST your prompt as JSON to <code>/chat</code>:</p>
	<pre>
	curl -X POST -H "Content-Type: application/json" \
	-d '{"prompt":"Hello, Llama!"}' \
	http://localhost:8080/chat
	</pre>
	`)
}

func main() {
	http.HandleFunc("/", homeHandler)           // Registers the homeHandler for the root endpoint
	http.HandleFunc("/chat", chatHandler)       // Registers the chatHandler for the /chat endpoint

	fmt.Println("Server listening on http://0.0.0.0:8080") // Prints a message indicating the server is running
	err := http.ListenAndServe("0.0.0.0:8080", nil)        // Starts the HTTP server on 0.0.0.0:8080
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
