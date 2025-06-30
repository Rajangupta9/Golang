package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type OllamaRequest struct {
	Model  string
	Prompt string
	Stream bool
}
type OllamaResponse struct {
	Response string
}
type userPrompt struct {
	Prompt string
	User   string
}
type User struct {
	User string
}

func ReadFile(filename string) string {
	data, err := os.ReadFile(filename + ".txt")
	if err != nil {
		return ""
	}
	return string(data)
}

func WriteFile(filename string, data string) bool {
	filename = filename + ".txt"
	prevData := ReadFile(filename)

	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	var finalData string
	if len(prevData) > 1 {
		finalData = prevData + "\n" + data
	} else {
		finalData = data
	}

	_, err = file.WriteString(finalData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return false
	}

	return true
}

// func ReadPDFAsString(filename string) (string, error) {
// 	fmt.Println(filename)
// 	f, err := os.Open(filename)
// 	if err != nil {
// 		return "", fmt.Errorf("error opening PDF file: %w", err)
// 	}
// 	defer f.Close()

// 	r, err := pdf.NewReader(f, -1)
// 	if err != nil {
// 		return "", fmt.Errorf("error reading PDF: %w", err)
// 	}

// 	var content strings.Builder
// 	numPages := r.NumPage()

// 	for i := 1; i <= numPages; i++ {
// 		page := r.Page(i)
// 		if page.V.IsNull() {
// 			continue
// 		}

// 		// Extract text from the page
// 		var buf strings.Builder
// 		pageContent := page.Content()
// 		for _, text := range pageContent.Text {
// 			buf.WriteString(text.S)
// 			buf.WriteString(" ")
// 		}

// 		content.WriteString(buf.String())
// 		content.WriteString("\n") // Add newline between pages
// 	}

// 	return content.String(), nil
// }

func handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Contol-Allow-Origin", "*")
	w.Header().Set("Access-Contol-Allow-Methos", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": Only POST method allowed}`, http.StatusMethodNotAllowed)
		return
	}

	var userInput userPrompt

	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		http.Error(w, "user input didn't decoded", http.StatusBadRequest)
		return
	}
	if userInput.Prompt == "" || userInput.User == "" {
		http.Error(w, "promp and user are empty", http.StatusBadRequest)
		return
	}
	filename := userInput.User
	data := ReadFile(filename)

	// if err != nil {
	// 	fmt.Println(err)
	// 	http.Error(w, "user file not read ", http.StatusInternalServerError)
	// 	return
	// }
	// fmt.Println(data + "this is data")

	// fmt.Println(data)

	template := ReadFile("template")

	// var jsonTemplate map[string]interface{}
	// json.Unmarshal([]byte(template), &jsonTemplate)

	// fmt.Println(jsonTemplate)
	fullPrompt := userInput.Prompt + data + template

	response, err := queryLLaMA(fullPrompt)
	if err != nil {
		http.Error(w, "ollama server error", http.StatusInternalServerError)
		return
	}

	fmt.Println(response)

	json.NewEncoder(w).Encode(response)

}

func queryLLaMA(fullPrompt string) (string, error) {
	// fmt.Println(fullPrompt)
	url := "http://localhost:11434/api/generate"

	requestBody, err := json.Marshal(OllamaRequest{
		Model:  "llama3",
		Prompt: fullPrompt,
		Stream: false,
	})

	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))

	if err != nil {
		return "", fmt.Errorf("error making request to ollama: %v", err)
	}
	defer resp.Body.Close()

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", fmt.Errorf("failed to decode ollama response: %v", err)
	}

	return ollamaResp.Response, nil
}

func userhandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Contol-Allow-Origin", "*")
	w.Header().Set("Access-Contol-Allow-Methos", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": Only POST method allowed}`, http.StatusMethodNotAllowed)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "user input didn't decoded", http.StatusBadRequest)
		return
	}
	if u.User == "" {
		http.Error(w, "user are empty", http.StatusBadRequest)
		return
	}
	filename := u.User + ".txt"

	data := ReadFile(filename)

	if data != "" {
		res, err := queryLLaMA(data + "this is the data of mine so its usefull so make only greating")

		if err != nil {
			http.Error(w, "lamma intenal server error", http.StatusInternalServerError)
			return
		} else {

			json.NewEncoder(w).Encode(res)
			return
		}

	}

	http.Error(w, "no user data", http.StatusBadRequest)

}

func main() {

	http.HandleFunc("/prompt", handle)
	http.HandleFunc("/user", userhandle)
	var Port string = "5000"
	fmt.Printf("sever start on port %s", Port)
	log.Fatal(http.ListenAndServe(":"+Port, nil))
}
