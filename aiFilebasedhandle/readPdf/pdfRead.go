// package readpdf

// import (
// 	"fmt"
// 	"os"
// 	"strings"


// 	"github.com/ledongthuc/pdf"
// )

// func ReadPDFAsString(filename string) (string, error) {
// 	fmt.Println("Reading file:", filename)
// 	if !strings.HasSuffix(filename, ".pdf") {
// 		filename=  "/home/rajan/Documents/learnGo/aiFilebasedhandle/readPdf/" +filename + ".pdf"
// 	}
// 	// Check if file exists
// 	if _, err := os.Stat(filename); os.IsNotExist(err) {
// 		return "", fmt.Errorf("file does not exist: %s", filename)
// 	}

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
// 	fmt.Printf("Total pages: %d\n", numPages)

// 	for i := 1; i <= numPages; i++ {
// 		fmt.Printf("Processing page %d...\n", i)
// 		page := r.Page(i)
// 		if page.V.IsNull() {
// 			fmt.Printf("Page %d is null, skipping\n", i)
// 			continue
// 		}

// 		// Extract text from the page
// 		var buf strings.Builder
// 		pageContent := page.Content()

// 		fmt.Printf("Page %d has %d text elements\n", i, len(pageContent.Text))

// 		for j, text := range pageContent.Text {
// 			if text.S != "" {
// 				buf.WriteString(text.S)
// 				buf.WriteString(" ")
// 				// Debug: print first few text elements
// 				if j < 5 {
// 					fmt.Printf("  Text element %d: '%s'\n", j, text.S)
// 				}
// 			}
// 		}

// 		pageText := buf.String()
// 		if pageText != "" {
// 			content.WriteString(pageText)
// 			content.WriteString("\n") // Add newline between pages
// 		}

// 		fmt.Printf("Page %d extracted %d characters\n", i, len(pageText))
// 	}

// 	result := content.String()
// 	fmt.Printf("Total extracted text length: %d characters\n", len(result))

// 	return result, nil
// }
package readPdf

import (
	"fmt"
	"os/exec"
)

func readPDFWithPython(filename string) (string, error) {
	cmd := exec.Command("python3", "extract.py", filename)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func ReadPDFAsString(filename string) string{
	text, err := readPDFWithPython(filename)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	
	return text
}