package readpdf

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// PDFReader handles PDF reading operations using only Go standard library
type PDFReader struct {
	basePath string
	verbose  bool
}

// NewPDFReader creates a new PDF reader with optional base path
func NewPDFReader(basePath string, verbose bool) *PDFReader {
	return &PDFReader{
		basePath: basePath,
		verbose:  verbose,
	}
}

// ReadPDFAsString reads a PDF file and returns its content as a string
func (pr *PDFReader) ReadPDFAsString(filename string) (string, error) {
	fullPath := pr.resolveFilePath(filename)

	if pr.verbose {
		fmt.Printf("Reading file: %s\n", fullPath)
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return "", fmt.Errorf("file does not exist: %s", fullPath)
	}

	// Read the entire PDF file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("error reading PDF file: %w", err)
	}

	// Check if it's a valid PDF
	if !bytes.HasPrefix(data, []byte("%PDF-")) {
		return "", fmt.Errorf("not a valid PDF file")
	}

	if pr.verbose {
		fmt.Printf("PDF file size: %d bytes\n", len(data))
	}

	return pr.extractTextFromPDF(data)
}

// resolveFilePath resolves the full file path
func (pr *PDFReader) resolveFilePath(filename string) string {
	// If it's already a full path or has .pdf extension, use as is
	if filepath.IsAbs(filename) || strings.HasSuffix(strings.ToLower(filename), ".pdf") {
		return filename
	}

	// If basePath is set, use it; otherwise use current directory
	basePath := pr.basePath
	if basePath == "" {
		basePath = "."
	}

	return filepath.Join(basePath, filename+".pdf")
}

// extractTextFromPDF extracts text from PDF data
func (pr *PDFReader) extractTextFromPDF(data []byte) (string, error) {
	var extractedText strings.Builder

	// Find all stream objects in the PDF
	streams := pr.findStreams(data)

	if pr.verbose {
		fmt.Printf("Found %d streams in PDF\n", len(streams))
	}

	for i, stream := range streams {
		if pr.verbose {
			fmt.Printf("Processing stream %d (length: %d bytes)\n", i+1, len(stream))
		}

		text := pr.extractTextFromStream(stream)
		if text != "" {
			extractedText.WriteString(text)
			extractedText.WriteString(" ")
		}
	}

	result := pr.cleanText(extractedText.String())

	if pr.verbose {
		fmt.Printf("Total extracted text length: %d characters\n", len(result))
	}

	return result, nil
}

// findStreams finds all stream objects in the PDF
func (pr *PDFReader) findStreams(data []byte) [][]byte {
	var streams [][]byte

	// Regular expression to find stream objects
	streamRegex := regexp.MustCompile(`stream\s*\n(.*?)\nendstream`)
	matches := streamRegex.FindAllSubmatch(data, -1)

	for _, match := range matches {
		if len(match) > 1 {
			streamData := match[1]

			// Try to decompress if it's compressed
			decompressed := pr.tryDecompress(streamData)
			if decompressed != nil {
				streams = append(streams, decompressed)
			} else {
				streams = append(streams, streamData)
			}
		}
	}

	return streams
}

// tryDecompress tries to decompress stream data
func (pr *PDFReader) tryDecompress(data []byte) []byte {
	// Try zlib/flate decompression
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil
	}

	return decompressed
}

// extractTextFromStream extracts readable text from a stream
func (pr *PDFReader) extractTextFromStream(stream []byte) string {
	var text strings.Builder

	// Convert to string for text processing
	streamStr := string(stream)

	// Look for text between parentheses (literal strings in PDF)
	textRegex := regexp.MustCompile(`\((.*?)\)`)
	matches := textRegex.FindAllStringSubmatch(streamStr, -1)

	for _, match := range matches {
		if len(match) > 1 {
			// Clean up the extracted text
			cleanText := pr.cleanPDFText(match[1])
			if cleanText != "" {
				text.WriteString(cleanText)
				text.WriteString(" ")
			}
		}
	}

	// Also look for hexadecimal strings
	hexRegex := regexp.MustCompile(`<([0-9A-Fa-f\s]+)>`)
	hexMatches := hexRegex.FindAllStringSubmatch(streamStr, -1)

	for _, match := range hexMatches {
		if len(match) > 1 {
			hexText := pr.hexToText(match[1])
			if hexText != "" {
				text.WriteString(hexText)
				text.WriteString(" ")
			}
		}
	}

	// Look for text after 'Tj' or 'TJ' operators
	tjRegex := regexp.MustCompile(`\((.*?)\)\s*Tj`)
	tjMatches := tjRegex.FindAllStringSubmatch(streamStr, -1)

	for _, match := range tjMatches {
		if len(match) > 1 {
			cleanText := pr.cleanPDFText(match[1])
			if cleanText != "" {
				text.WriteString(cleanText)
				text.WriteString(" ")
			}
		}
	}

	return text.String()
}

// cleanPDFText cleans up text extracted from PDF
func (pr *PDFReader) cleanPDFText(text string) string {
	// Handle escape sequences
	text = strings.ReplaceAll(text, "\\n", "\n")
	text = strings.ReplaceAll(text, "\\r", "\r")
	text = strings.ReplaceAll(text, "\\t", "\t")
	text = strings.ReplaceAll(text, "\\(", "(")
	text = strings.ReplaceAll(text, "\\)", ")")
	text = strings.ReplaceAll(text, "\\\\", "\\")

	// Remove control characters but keep printable ones
	var cleaned strings.Builder
	for _, r := range text {
		if r >= 32 && r < 127 || r == '\n' || r == '\r' || r == '\t' {
			cleaned.WriteRune(r)
		}
	}

	return cleaned.String()
}

// hexToText converts hexadecimal string to text
func (pr *PDFReader) hexToText(hexStr string) string {
	// Remove spaces
	hexStr = strings.ReplaceAll(hexStr, " ", "")

	// Must be even length
	if len(hexStr)%2 != 0 {
		return ""
	}

	var result strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		if i+1 < len(hexStr) {
			hexByte := hexStr[i : i+2]
			if val, err := strconv.ParseUint(hexByte, 16, 8); err == nil {
				if val >= 32 && val < 127 { // Printable ASCII
					result.WriteByte(byte(val))
				}
			}
		}
	}

	return result.String()
}

// cleanText performs final cleanup of extracted text
func (pr *PDFReader) cleanText(text string) string {
	// Remove multiple spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")

	// Remove multiple newlines
	newlineRegex := regexp.MustCompile(`\n\s*\n`)
	text = newlineRegex.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text)
}

// Simple PDF text extractor that attempts to extract text from any PDF
func (pr *PDFReader) extractTextSimple(data []byte) string {
	var result strings.Builder

	// Convert to string for regex processing
	content := string(data)

	// Find all text strings in the PDF (very basic approach)
	// This looks for patterns like (text) or <hextext>
	patterns := []string{
		`\(([^)]+)\)`,        // (text)
		`<([0-9A-Fa-f\s]+)>`, // <hextext>
	}

	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllStringSubmatch(content, -1)

		for _, match := range matches {
			if len(match) > 1 {
				if strings.Contains(pattern, "Fa-f") {
					// Hex pattern
					text := pr.hexToText(match[1])
					if len(text) > 0 {
						result.WriteString(text + " ")
					}
				} else {
					// Regular text pattern
					text := pr.cleanPDFText(match[1])
					if len(text) > 0 {
						result.WriteString(text + " ")
					}
				}
			}
		}
	}

	return pr.cleanText(result.String())
}

// Convenience functions
func ReadPDFAsString(filename string) (string, error) {
	reader := NewPDFReader("", false)
	return reader.ReadPDFAsString(filename)
}

func ReadPDFAsStringVerbose(filename string) (string, error) {
	reader := NewPDFReader("", true)
	return reader.ReadPDFAsString(filename)
}

func ReadPDFFromPath(basePath, filename string) (string, error) {
	reader := NewPDFReader(basePath, false)
	return reader.ReadPDFAsString(filename)
}

func ReadPDFFromPathVerbose(basePath, filename string) (string, error) {
	reader := NewPDFReader(basePath, true)
	return reader.ReadPDFAsString(filename)
}
