// package handlers

// import (
// 	"bytes"
// 	"crypto/sha256"
// 	"encoding/hex"
// 	"fmt"
// 	"html/template"
// 	"net/http"
// 	"os"
// 	"strconv"
// 	"strings"
// )

// type AsciiArtData struct {
// 	Text     string
// 	AsciiArt string
// 	Banner   string
// }

// // AsciiArtHandler handles POST requests, processes input text,
// // generates ASCII art, and renders an HTML template with the result.
// func AsciiArtHandler(w http.ResponseWriter, r *http.Request) {
// 	if r.URL.Path != "/ascii-art" {
// 		ErrorHandler(w, "Page not found", http.StatusNotFound)
// 		return
// 	}

// 	switch r.Method {
// 	case "POST":
// 		// Get input text and banner from the request.
// 		text := r.FormValue("text")
// 		banner := r.FormValue("banner")
// 		download := r.FormValue("download")

// 		// Validate input
// 		if text == "" || banner == "" {
// 			ErrorHandler(w, "Bad request", http.StatusBadRequest)
// 			return
// 		}

// 		// Read the banner file and generate ASCII art
// 		lines, err := readBanner(banner)
// 		if err != nil {
// 			ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
// 			return
// 		}

// 		// Split the input text into multiple lines
// 		textLines := strings.Split(text, "\r\n")

// 		// Process the input text and generate ASCII art.
// 		var asciiArtBuffer bytes.Buffer
// 		for _, words := range textLines {
// 			for i := 0; i < 8; i++ {
// 				for _, char := range words {
// 					if !(char >= 32 && char <= 126) {
// 						ErrorHandler(w, "Bad request", http.StatusBadRequest)
// 						return
// 					}
// 					asciiArtBuffer.WriteString(lines[int(char-' ')*9+1+i] + " ")
// 				}
// 				asciiArtBuffer.WriteString("\n")
// 			}
// 			asciiArtBuffer.WriteString("\n")
// 		}

// 		asciiArt := asciiArtBuffer.String()

// 		if download == "true" {
// 			// Set headers for file download
// 			w.Header().Set("Content-Disposition", "attachment; filename=ascii_art.txt")
// 			w.Header().Set("Content-Type", "text/plain")
// 			w.Header().Set("Content-Length", strconv.Itoa(len(asciiArt)))

// 			// Write the ASCII art to the response
// 			_, err := w.Write([]byte(asciiArt))
// 			if err != nil {
// 				ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
// 				return
// 			}
// 			return
// 		}

// 		// Render the ASCII art template
// 		data := AsciiArtData{Text: text, AsciiArt: asciiArt, Banner: banner}
// 		tmpl, err := template.ParseFiles("../templates/asciiart.html")
// 		if err != nil {
// 			ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
// 			return
// 		}

// 		err = tmpl.Execute(w, data)
// 		if err != nil {
// 			ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
// 		}
// 	default:
// 		ErrorHandler(w, "Method not allowed", http.StatusMethodNotAllowed)
// 	}
// }

// // readBanner reads the banner file content and returns the content as a slice of strings.
// func readBanner(banner string) ([]string, error) {
// 	path := "../banners/" + banner
// 	data, err := os.ReadFile(path)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Check file's integrity.
// 	hashFile := checkSum(data)
// 	hashShadow := "26b94d0b134b77e9fd23e0360bfd81740f80fb7f6541d1d8c5d85e73ee550f73"
// 	hashStandard := "e194f1033442617ab8a78e1ca63a2061f5cc07a3f05ac226ed32eb9dfd22a6bf"
// 	hashThinkertoy := "092d0cde973bfbb02522f18e00e8612e269f53bac358bb06f060a44abd0dbc52"

// 	if hashFile != hashShadow && hashFile != hashStandard && hashFile != hashThinkertoy {
// 		return nil, fmt.Errorf("file corrupted")
// 	}

// 	lines := strings.Split(string(data), "\n")
// 	return lines, nil
// }

// // checkSum calculates the SHA-256 checksum of a given byte slice and returns it as a hexadecimal string.
// func checkSum(data []byte) string {
// 	hash := sha256.Sum256(data)
// 	return hex.EncodeToString(hash[:])
// }

package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type AsciiArtData struct {
	Text     string
	AsciiArt string
	Banner   string
}

// AsciiArtHandler handles POST requests, process input text,
// generate ASCII art and render an HTML template with result.
func AsciiArtHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		// Ensure the banners directory exists
		err := os.MkdirAll("../banners", os.ModePerm)
		if err != nil {
			log.Println("Error creating banners directory:", err)
			ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// Get input text and banner from request.
		text := r.FormValue("text")
		banner := r.FormValue("banner")
		download := r.FormValue("download")
		// Validate input
		if text == "" || banner == "" {
			log.Println("Invalid input")
			ErrorHandler(w, "Bad request", http.StatusBadRequest)
			return
		}
		if banner != "standard.txt" && banner != "shadow.txt" && banner != "thinkertoy.txt" {
			log.Println("Invalid banner")
			ErrorHandler(w, "Invalid banner", http.StatusInternalServerError)
			return
		}
		// Read the banner file and generate ASCII art
		lines, err := readBanner(banner)
		if err != nil {
			log.Printf("Error reading banner: %v", err)
			log.Println("Initializing banner download")
			// Execute the command to download missing files
			err = downloadBannerFile(banner)
			if err != nil {
				log.Println("Error downloading file:", err)
				ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			log.Print("File downloaded successfully")
			lines, err = readBanner(banner) // Read the new banner file
			if err != nil {
				log.Println("Error reading banner after downloading:", err)
				ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
				return
			}
		}
		// Split the input text into multiple lines
		textLines := strings.Split(text, "\r\n")
		// Process the input text and generate ASII art.
		var asciiArtBuffer bytes.Buffer
		for _, words := range textLines {
			for i := 0; i < 8; i++ {
				for _, char := range words {
					if !(char >= 32 && char <= 126) {
						log.Println("Invalid character")
						ErrorHandler(w, "Bad request", http.StatusBadRequest)
						return
					}
					asciiArtBuffer.WriteString(lines[int(char-' ')*9+1+i])
				}
				asciiArtBuffer.WriteString("\n")
			}
			asciiArtBuffer.WriteString("\n")
		}
		asciiArt := asciiArtBuffer.String()

		if download == "true" {
			// Set headers for file download
			w.Header().Set("Content-Disposition", "attachment; filename=ascii_art.txt")
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", strconv.Itoa(len(asciiArt)))

			// Write the ASCII art to the response
			_, err := w.Write([]byte(asciiArt))
			if err != nil {
				ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			return
		}
		// Render the ASCII art template
		data := AsciiArtData{Text: text, AsciiArt: asciiArt, Banner: banner}
		tmpl, err := template.ParseFiles("../templates/asciiart.html")
		if err != nil {
			log.Println("Error parsing ASCII art template")
			ErrorHandler(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Println("Error executing ASCII art template")
			ErrorHandler(w, "Error Internal server error", http.StatusInternalServerError)
		}
	default:
		log.Println("Invalid request method")
		ErrorHandler(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ReadBanner reads banner file content and returns splited content as a slice of strings.
func readBanner(banner string) ([]string, error) {
	path := "../banners/" + banner
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Check file's integrity.
	hashFile := checkSum(data)
	hashShadow := "26b94d0b134b77e9fd23e0360bfd81740f80fb7f6541d1d8c5d85e73ee550f73"
	hashStandard := "e194f1033442617ab8a78e1ca63a2061f5cc07a3f05ac226ed32eb9dfd22a6bf"
	hashThinkertoy := "092d0cde973bfbb02522f18e00e8612e269f53bac358bb06f060a44abd0dbc52"
	if hashFile != hashShadow && hashFile != hashStandard && hashFile != hashThinkertoy {
		return nil, fmt.Errorf("file corrupted")
	}
	lines := strings.Split(string(data), "\n")
	return lines, nil
}

// CheckSum calculates the SHA-256 checksum of a given byte slice and returns it as a hexadecimal string.
func checkSum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// DownloadBannerFile downloads a banner file from a given URL and saves it to the local directory.
func downloadBannerFile(banner string) error {
	url := "https://raw.githubusercontent.com/stkisengese/ascii-art-server/master/banners/" + banner
	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download banner file: %w", err)
	}
	defer resp.Body.Close()
	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download banner file: received status code %d", resp.StatusCode)
	}
	// Create the local file to save the banner
	filePath := "../banners/" + banner
	out, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create banner file: %w", err)
	}
	defer out.Close()
	// Copy the downloaded content to the local file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write banner file: %w", err)
	}
	return nil
}
