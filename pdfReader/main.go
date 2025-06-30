// package main

// import (
// 	"fmt"
// 	"log"

// 	"github.com/unidoc/unipdf/v3/common/license"
// 	"github.com/unidoc/unipdf/v3/extractor"
// 	"github.com/unidoc/unipdf/v3/model"
// )

// func init() {
// 	// Free tier license (required to enable extraction)
// 	err := license.SetMeteredKey("45178fd5c7490bcae1a2bf9710e37267d235c1caa056cc5fbfe43ceec7b40538") // ← Replace with your key
// 	if err != nil {
// 		log.Fatalf("License error: %v", err)
// 	}
// }

// func ReadPDFText(filename string) string {
// 	pdfReader, closer, err := model.NewPdfReaderFromFile(filename, nil)
// 	if err != nil {
// 		log.Fatalf("Failed to open PDF file: %v", err)
// 	}
// 	if closer != nil {
// 		defer closer.Close()
// 	}

// 	numPages, err := pdfReader.GetNumPages()
// 	if err != nil {
// 		log.Fatalf("Failed to get page count: %v", err)
// 	}

// 	var fullText string
// 	for i := 1; i <= numPages; i++ {
// 		page, err := pdfReader.GetPage(i)
// 		if err != nil {
// 			log.Fatalf("Error getting page %d: %v", i, err)
// 		}
// 		ex, err := extractor.New(page)
// 		if err != nil {
// 			log.Fatalf("Error creating extractor: %v", err)
// 		}
// 		text, err := ex.ExtractText()
// 		if err != nil {
// 			log.Fatalf("Error extracting text: %v", err)
// 		}
// 		fullText += text + "\n"
// 	}

// 	return fullText
// }

// func main() {
// 	text := ReadPDFText("sample.pdf")
// 	fmt.Println(text)
// }


package main

import (
	"fmt"
	"os/exec"
)

func ReadPDFWithPython(filename string) (string, error) {
	cmd := exec.Command("python3", "extract.py", filename)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func main() {
	text, err := ReadPDFWithPython("sample.pdf")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Extracted PDF Content:\n")
	fmt.Println(text)
}


// package main

// import (
// 	"fmt"
// 	"log"
// 	"os"

// 	"rsc.io/pdf"
// )

// func ReadPDFAsString(filename string) string {
// 	f, err := os.Open(filename)
// 	if err != nil {
// 		log.Fatalf("Error opening PDF file: %v", err)
// 	}
// 	defer f.Close()

// 	r, err := pdf.NewReader(f, -1)
// 	if err != nil {
// 		log.Fatalf("Error reading PDF: %v", err)
// 	}

// 	var text string
// 	for i := 1; i <= r.NumPage(); i++ {
// 		page := r.Page(i)
// 		if page.V.IsNull() {
// 			continue
// 		}
// 		content := page.Content()
// 		for _, textObj := range content.Text {
// 			text += textObj.S
// 		}
// 	}

// 	return text
// }

// func main() {
// 	text := ReadPDFAsString("sample.pdf")
// 	fmt.Println(text)
// }
