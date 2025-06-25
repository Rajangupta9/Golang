package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

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

type User struct {
	User string `json:"user"`
}

type ComponentResult struct {
	ComponentType string      `json:"component_type"`
	Data          interface{} `json:"data"`
	Error         error       `json:"error,omitempty"`
}

func ReadFile(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}
	return string(data)
}

func WriteFile(filename string, data string) bool {
	prevData := ReadFile(filename)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return false
	}
	defer file.Close()

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

func setupCORS(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func getComponentTemplates() map[string]string {
	// Profile component template
	profileTemplate := `{
		"component": "profile",
		"pr_img": "/images/digitalCard/dbcv2/profile_1.webp",
		"br_img": "/images/digitalCard/dbcv2/barand_logo_9.webp",
		"name": "Name",
		"desc": "Title",
		"company": "Company",
		"contact_shortcuts": [
			{
				"type": "mobile",
				"value": "0000000000"
			},
			{
				"type": "email",
				"value": "youremail@domain.com"
			},
			{
				"type": "sms",
				"value": "0000000000"
			}
		]
	}`

	// About component template
	aboutTemplate := `{
		"component": "text_desc",
		"title": "About Me",
		"desc": "Description",
		"title_config": {
			"bold": 1,
			"italic": 0,
			"align": "center",
			"lock": "unlock"
		},
		"desc_config": {
			"bold": 0,
			"italic": 0,
			"align": "center",
			"lock": "unlock"
		}
	}`

	// Contact component template
	contactTemplate := `{
		"component": "contact",
		"contact_title": "Contact Us",
		"icon_img": "/images/digitalCard/contactus.png",
		"floating_button_label": "Add to Contact",
		"ebusiness_card_enable": 1,
		"contact_infos": [
			{
				"type": "number",
				"title": "Call Us",
				"label": "Mobile ",
				"number": "123 456 7890"
			},
			{
				"type": "email",
				"title": "Email",
				"label": "Email ",
				"email": "contactme@domain.com"
			},
			{
				"type": "address",
				"title": "Address",
				"street": "Street",
				"city": "City",
				"country": "Country",
				"state": "State",
				"zip": "Zipcode",
				"action_button_label": "Direction",
				"action_button_link": "#"
			}
		]
	}`

	// Social links component template
	socialTemplate := `{
		"component": "social_link",
		"title": "Social Links",
		"desc": "Description",
		"title_config": {
			"bold": 1,
			"italic": 0,
			"align": "center",
			"lock": "unlock"
		},
		"desc_config": {
			"bold": 0,
			"italic": 0,
			"align": "center",
			"lock": "unlock"
		},
		"links": [
			{
				"type": "facebook",
				"url": "",
				"title": "Facebook",
				"subtitle": "Follow us on Facebook",
				"icon_img": "/images/digitalCard/fb_icon@72x.png"
			},
			{
				"type": "instagram",
				"url": "",
				"title": "Instagram",
				"subtitle": "Follow us on Instagram",
				"icon_img": "/images/digitalCard/insta_icon@72x.png"
			},
			{
				"type": "twitter",
				"url": "",
				"title": "Twitter",
				"subtitle": "Follow us on Twitter",
				"icon_img": "/images/digitalCard/tw_icon@72x.png"
			},
			{
				"type": "linkedin",
				"url": "",
				"title": "LinkedIn",
				"subtitle": "Follow us on LinkedIn",
				"icon_img": "/images/digitalCard/linkedin_icon@72x.png"
			},
			{
				"type": "github",
				"url": "",
				"title": "GitHub",
				"subtitle": "Follow us on GitHub",
				"icon_img": "/images/digitalCard/github_icon@72x.png"
			}
		]
	}`

	return map[string]string{
		"profile": profileTemplate,
		"about":   aboutTemplate,
		"contact": contactTemplate,
		"social":  socialTemplate,
	}
}

func createComponentPrompt(componentType string, componentTemplate string, userData string, userPrompt string) string {
	basePrompt := fmt.Sprintf(`
USER DATA:
%s

USER REQUEST:
%s

TEMPLATE TO FILL:
%s

INSTRUCTIONS:
- Extract relevant information from the USER DATA above
- Fill the template with extracted data
- Replace placeholder values with actual data from user information
- Keep the exact JSON structure provided in template
- If data is not available, keep the placeholder values
- Return ONLY the JSON object, no additional text
- For component type '%s', focus on:`, userData, userPrompt, componentTemplate, componentType)

	switch componentType {
	case "profile":
		basePrompt += `
  * Extract: name, job title/position, company name, phone number, email
  * Update: name, desc (job title), company, contact_shortcuts values`
	case "about":
		basePrompt += `
  * Extract: professional summary, bio, description, key skills overview
  * Update: desc field with professional summary`
	case "contact":
		basePrompt += `
  * Extract: phone numbers, emails, address, location details
  * Update: contact_infos array with actual contact information`
	case "social":
		basePrompt += `
  * Extract: LinkedIn, GitHub, Facebook, Instagram, Twitter URLs
  * Update: links array with actual social media URLs`
	}

	return basePrompt
}

func processComponentWithTemplate(componentType string, userData string, userPrompt string, results chan<- ComponentResult, wg *sync.WaitGroup) {
	defer wg.Done()

	componentTemplates := getComponentTemplates()
	template, exists := componentTemplates[componentType]
	if !exists {
		results <- ComponentResult{
			ComponentType: componentType,
			Data:          nil,
			Error:         fmt.Errorf("template not found for component: %s", componentType),
		}
		return
	}

	fullPrompt := createComponentPrompt(componentType, template, userData, userPrompt)
	
	// Limit prompt size for ollama 3.2:1b - keep it smaller
	if len(fullPrompt) > 1200 {
		// Truncate user data but keep template and instructions
		maxUserData := 400
		if len(userData) > maxUserData {
			userData = userData[:maxUserData] + "..."
		}
		fullPrompt = createComponentPrompt(componentType, template, userData, userPrompt)
	}

	response, err := queryLLaMA(fullPrompt)
	if err != nil {
		results <- ComponentResult{
			ComponentType: componentType,
			Data:          nil,
			Error:         err,
		}
		return
	}

	// Try to parse the JSON response
	var jsonResponse interface{}
	if err := json.Unmarshal([]byte(response), &jsonResponse); err != nil {
		// If JSON parsing fails, return the raw response
		fmt.Printf("Warning: Failed to parse JSON for %s component: %v\n", componentType, err)
		results <- ComponentResult{
			ComponentType: componentType,
			Data:          response, // Return raw response
			Error:         nil,
		}
		return
	}

	results <- ComponentResult{
		ComponentType: componentType,
		Data:          jsonResponse,
		Error:         nil,
	}
}

func processAllComponents(userData string, userPrompt string) (map[string]interface{}, error) {
	componentTypes := []string{"profile", "about", "contact", "social"}
	results := make(chan ComponentResult, len(componentTypes))
	var wg sync.WaitGroup

	// Process each component concurrently
	for _, componentType := range componentTypes {
		wg.Add(1)
		go processComponentWithTemplate(componentType, userData, userPrompt, results, &wg)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	processedComponents := make(map[string]interface{})
	for result := range results {
		if result.Error != nil {
			fmt.Printf("Error processing %s: %v\n", result.ComponentType, result.Error)
			continue
		}
		processedComponents[result.ComponentType] = result.Data
	}

	return processedComponents, nil
}

func buildFinalResponse(processedComponents map[string]interface{}) map[string]interface{} {
	// Load the base template
	templateData := ReadFile("template.json")
	var baseTemplate map[string]interface{}
	
	if err := json.Unmarshal([]byte(templateData), &baseTemplate); err != nil {
		fmt.Printf("Error parsing base template: %v\n", err)
		// Create a basic structure if template parsing fails
		return map[string]interface{}{
			"template_id": "b_685bb94064ea9b78d567cade",
			"qr_codes": []interface{}{
				map[string]interface{}{
					"qr_name":   "",
					"short_url": "",
					"content":   buildContentArray(processedComponents),
				},
			},
		}
	}

	// Update the content array in the template
	if qrCodes, ok := baseTemplate["qr_codes"].([]interface{}); ok && len(qrCodes) > 0 {
		if qrCode, ok := qrCodes[0].(map[string]interface{}); ok {
			qrCode["content"] = buildContentArray(processedComponents)
			qrCodes[0] = qrCode
			baseTemplate["qr_codes"] = qrCodes
		}
	}

	return baseTemplate
}

func buildContentArray(processedComponents map[string]interface{}) []interface{} {
	var contentArray []interface{}

	// Add components in specific order
	componentOrder := []string{"profile", "about", "contact", "social"}
	
	for _, componentType := range componentOrder {
		if componentData, exists := processedComponents[componentType]; exists {
			contentArray = append(contentArray, componentData)
		}
	}

	// Add remaining static components from template that weren't processed
	staticComponents := getStaticComponents()
	contentArray = append(contentArray, staticComponents...)

	return contentArray
}

func getStaticComponents() []interface{} {
	// Return static components that don't need AI processing
	return []interface{}{
		map[string]interface{}{
			"component": "images",
			"title":     "",
			"desc":      "",
			"title_config": map[string]interface{}{
				"bold":   1,
				"italic": 0,
				"align":  "center",
				"lock":   "unlock",
			},
			"desc_config": map[string]interface{}{
				"bold":   0,
				"italic": 0,
				"align":  "center",
				"lock":   "unlock",
			},
			"view_type": "list",
			"images": []string{
				"/images/digitalCard/image_1.png",
				"/images/digitalCard/image_2.png",
			},
		},
		map[string]interface{}{
			"component": "web_links",
			"title":     "Web Links",
			"desc":      "Description",
			"title_config": map[string]interface{}{
				"bold":   1,
				"italic": 0,
				"align":  "center",
				"lock":   "unlock",
			},
			"desc_config": map[string]interface{}{
				"bold":   0,
				"italic": 0,
				"align":  "center",
				"lock":   "unlock",
			},
			"links": []interface{}{
				map[string]interface{}{
					"url":      "https://www.mycoolbrand.com",
					"title":    "Title",
					"subtitle": "Sub Title",
					"icon_img": "/images/digitalCard/weblink.png",
				},
			},
		},
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	setupCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Only POST method allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var userInput UserPrompt
	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		http.Error(w, `{"error": "Invalid JSON input"}`, http.StatusBadRequest)
		return
	}

	if userInput.Prompt == "" || userInput.User == "" {
		http.Error(w, `{"error": "Prompt and user fields are required"}`, http.StatusBadRequest)
		return
	}

	filename := userInput.User + ".txt"
	userData := ReadFile(filename)

	if userData == "" {
		http.Error(w, `{"error": "No user data found"}`, http.StatusNotFound)
		return
	}

	// Process all components with full user data and template
	processedComponents, err := processAllComponents(userData, userInput.Prompt)
	if err != nil {
		http.Error(w, `{"error": "Failed to process components"}`, http.StatusInternalServerError)
		return
	}

	// Build final response using the template structure
	finalResponse := buildFinalResponse(processedComponents)

	// Store the updated user data
	if userInput.Prompt != "" {
		WriteFile(filename, "Updated with: "+userInput.Prompt)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(finalResponse)
}

func queryLLaMA(prompt string) (string, error) {
	url := "http://localhost:11434/api/generate"

	requestBody, err := json.Marshal(OllamaRequest{
		Model:  "llama3.2:1b",
		Prompt: prompt,
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
	setupCORS(w)

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "Only POST method allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var u User
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, `{"error": "Invalid JSON input"}`, http.StatusBadRequest)
		return
	}

	if u.User == "" {
		http.Error(w, `{"error": "User field is required"}`, http.StatusBadRequest)
		return
	}

	filename := u.User + ".txt"
	data := ReadFile(filename)

	if data != "" {
		greetingPrompt := fmt.Sprintf(`
Based on this user information, create a personalized professional greeting:

USER DATA:
%s

Create a brief, professional greeting message (max 2-3 sentences) that acknowledges their background and expertise. Return only the greeting text, no additional formatting.`, data)

		// Limit prompt size for ollama 3.2:1b
		if len(greetingPrompt) > 800 {
			// Truncate data but keep instruction
			if len(data) > 400 {
				data = data[:400] + "..."
			}
			greetingPrompt = fmt.Sprintf(`
Based on this user information, create a personalized professional greeting:

USER DATA:
%s

Create a brief, professional greeting message (max 2-3 sentences). Return only the greeting text.`, data)
		}

		res, err := queryLLaMA(greetingPrompt)
		if err != nil {
			http.Error(w, `{"error": "LLaMA internal server error"}`, http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"greeting": res,
			"user":     u.User,
		}
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, `{"error": "No user data found"}`, http.StatusNotFound)
}

func main() {
	http.HandleFunc("/prompt", handle)
	http.HandleFunc("/user", userhandle)
	
	port := "5000"
	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}