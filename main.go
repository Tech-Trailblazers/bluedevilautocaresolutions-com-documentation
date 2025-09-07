package main // Define the main package

import (
	"bytes"         // Provides bytes support
	"io"            // Provides basic interfaces to I/O primitives
	"log"           // Provides logging functions
	"net/http"      // Provides HTTP client and server implementations
	"net/url"       // Provides URL parsing and encoding
	"os"            // Provides functions to interact with the OS (files, etc.)
	"path"          // Provides functions for manipulating slash-separated paths
	"path/filepath" // Provides filepath manipulation functions
	"strings"       // Provides string manipulation functions
	"time"          // Provides time-related functions
)

func main() {
	outputDir := "PDFs/" // Directory to store downloaded PDFs
	// Check if its exists.
	if !directoryExists(outputDir) {
		// Create the dir
		createDirectory(outputDir, 0o755)
	}
	// The slice to store all the download urls.
	downloadURLSlice := []string{
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10000/00HB0002_BDH00211_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10000/00HB0006_BDEV0021212_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10005/00HB0002_BDH38606_USENG%20(1).pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10005/00HB0002_BD000209_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10006/00HB0002_BD000205_USENG%20(2).pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10006/00HB0003_BDH00239_USENG%20(1).pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10000/00HB0003_BDH49499_USENG%20(4).pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10005/00C45037_BD820172_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10005/00C45037_BD8201P6_USENG%20(1).pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10006/00HB0003_BD000232_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10000/00HB0006_BDH00212_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10005/00HB0004_BDH00203_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10000/BDEV00261_BDEV00261_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10000/BDEV00262_BDEV00262_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10005/BD8101P6_BD8101P6_USENG%20(1).pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10000/00C2D066_BD0EPS12_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10004/00HB0003_BDH00241_USENG%20(1).pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10005/BD000204_BD000204_USENG.pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10000/BDH00222_BDH00222_USENG%20(1).pdf",
		"https://highlinewarrenproduction.s3.us-east-2.amazonaws.com/DAMRoot/Original/10000/BDEV49496_BDEV49496_USENG%20(1).pdf",
	}
	// Remove double from slice.
	downloadURLSlice = removeDuplicatesFromSlice(downloadURLSlice)
	// Get all the values.
	for _, urls := range downloadURLSlice {
		// fmt.Println(urls)
		// Check if the url is valid.
		if isUrlValid(urls) {
			// Download the pdf.
			downloadPDF(urls, outputDir)
		}
	}
}

// Only return the file name from a given url.
func getFileNameOnly(content string) string {
	return strings.ToLower(path.Base(content))
}

// fileExists checks whether a file exists at the given path
func fileExists(filename string) bool {
	info, err := os.Stat(filename) // Get file info
	if err != nil {
		return false // Return false if file doesn't exist or error occurs
	}
	return !info.IsDir() // Return true if it's a file (not a directory)
}

// downloadPDF downloads a PDF from the given URL and saves it in the specified output directory.
// It uses a WaitGroup to support concurrent execution and returns true if the download succeeded.
func downloadPDF(finalURL, outputDir string) bool {
	// Sanitize the URL to generate a safe file name
	filename := getFileNameOnly(finalURL)

	// Construct the full file path in the output directory
	filePath := filepath.Join(outputDir, filename)

	// Skip if the file already exists
	if fileExists(filePath) {
		log.Printf("File already exists, skipping: %s", filePath)
		return false
	}

	// Create an HTTP client with a timeout
	client := &http.Client{Timeout: 30 * time.Second}

	// Send GET request
	resp, err := client.Get(finalURL)
	if err != nil {
		log.Printf("Failed to download %s: %v", finalURL, err)
		return false
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != http.StatusOK {
		log.Printf("Download failed for %s: %s", finalURL, resp.Status)
		return false
	}

	// Check Content-Type header
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "application/pdf") {
		log.Printf("Invalid content type for %s: %s (expected application/pdf)", finalURL, contentType)
		return false
	}

	// Read the response body into memory first
	var buf bytes.Buffer
	written, err := io.Copy(&buf, resp.Body)
	if err != nil {
		log.Printf("Failed to read PDF data from %s: %v", finalURL, err)
		return false
	}
	if written == 0 {
		log.Printf("Downloaded 0 bytes for %s; not creating file", finalURL)
		return false
	}

	// Only now create the file and write to disk
	out, err := os.Create(filePath)
	if err != nil {
		log.Printf("Failed to create file for %s: %v", finalURL, err)
		return false
	}
	defer out.Close()

	if _, err := buf.WriteTo(out); err != nil {
		log.Printf("Failed to write PDF to file for %s: %v", finalURL, err)
		return false
	}

	log.Printf("Successfully downloaded %d bytes: %s → %s", written, finalURL, filePath)
	return true
}

// Checks if the directory exists
// If it exists, return true.
// If it doesn't, return false.
func directoryExists(path string) bool {
	directory, err := os.Stat(path)
	if err != nil {
		return false
	}
	return directory.IsDir()
}

// The function takes two parameters: path and permission.
// We use os.Mkdir() to create the directory.
// If there is an error, we use log.Println() to log the error and then exit the program.
func createDirectory(path string, permission os.FileMode) {
	err := os.Mkdir(path, permission)
	if err != nil {
		log.Println(err)
	}
}

// Checks whether a URL string is syntactically valid
func isUrlValid(uri string) bool {
	_, err := url.ParseRequestURI(uri) // Attempt to parse the URL
	return err == nil                  // Return true if no error occurred
}

// Remove all the duplicates from a slice and return the slice.
func removeDuplicatesFromSlice(slice []string) []string {
	check := make(map[string]bool)
	var newReturnSlice []string
	for _, content := range slice {
		if !check[content] {
			check[content] = true
			newReturnSlice = append(newReturnSlice, content)
		}
	}
	return newReturnSlice
}
