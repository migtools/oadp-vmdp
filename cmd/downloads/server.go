// Package main implements the oadp-vmdp download server for binary distribution.
package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	binaryPrefix     = "oadp-vmdp_"
	licenseFile      = "LICENSE"
	bytesPerMB       = 1024 * 1024
	minPlatformParts = 3
	readTimeout      = 10 * time.Second
	writeTimeout     = 60 * time.Second
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var (
	binaryDir    = getEnv("ARCHIVE_DIR", "/archives")
	port         = getEnv("PORT", "8080")
	pageTemplate = template.Must(template.ParseFS(templateFS, "templates/index.html"))
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

type binaryFile struct {
	Name     string
	Size     float64
	OS       string
	Arch     string
	Checksum string
}

func main() {
	files, err := discoverBinaries()
	if err != nil || len(files) == 0 {
		log.Fatal("No binaries found in ", binaryDir)
	}

	log.Printf("Found %d binaries", len(files))

	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal("Failed to load static files: ", err)
	}

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))
	http.HandleFunc("/", listBinaries)
	http.HandleFunc("/download/", downloadBinary)

	log.Printf("Starting server on port %s", port)
	log.Printf("Serving binaries from %s", binaryDir)

	srv := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// discoverBinaries finds oadp-vmdp binaries (excluding .sha256 and LICENSE files).
func discoverBinaries() ([]string, error) {
	entries, err := os.ReadDir(binaryDir)
	if err != nil {
		return nil, fmt.Errorf("reading binary directory: %w", err)
	}

	var binaries []string

	for _, e := range entries {
		name := e.Name()

		if e.IsDir() || strings.HasSuffix(name, ".sha256") || name == licenseFile {
			continue
		}

		if strings.HasPrefix(name, binaryPrefix) {
			binaries = append(binaries, filepath.Join(binaryDir, name))
		}
	}

	return binaries, nil
}

// readChecksum reads a SHA256 checksum from a per-file .sha256 sidecar file.
func readChecksum(filePath string) string {
	data, err := os.ReadFile(filePath + ".sha256") //nolint:gosec // path is constructed from binaryDir constant
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(data))
	if len(fields) > 0 {
		return fields[0]
	}

	return ""
}

// parsePlatform extracts OS and architecture from a filename like
// oadp-vmdp_linux_amd64 or oadp-vmdp_windows_arm64.exe.
func parsePlatform(filename string) (string, string) {
	name := strings.TrimSuffix(filename, ".exe")

	parts := strings.Split(name, "_")
	if len(parts) >= minPlatformParts {
		return parts[len(parts)-2], parts[len(parts)-1]
	}

	return "unknown", "unknown"
}

func listBinaries(w http.ResponseWriter, _ *http.Request) {
	files, err := discoverBinaries()
	if err != nil {
		http.Error(w, "Error listing binaries", http.StatusInternalServerError)
		return
	}

	hasLicense := false
	if _, err := os.Stat(filepath.Join(binaryDir, licenseFile)); err == nil {
		hasLicense = true
	}

	var linuxFiles, windowsFiles []binaryFile

	for _, file := range files {
		name := filepath.Base(file)

		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		size := float64(info.Size()) / bytesPerMB
		osName, arch := parsePlatform(name)
		checksum := readChecksum(file)

		bf := binaryFile{Name: name, Size: size, OS: osName, Arch: arch, Checksum: checksum}

		switch osName {
		case "linux":
			linuxFiles = append(linuxFiles, bf)
		case "windows":
			windowsFiles = append(windowsFiles, bf)
		}
	}

	data := struct {
		LinuxFiles   []binaryFile
		WindowsFiles []binaryFile
		HasLicense   bool
	}{linuxFiles, windowsFiles, hasLicense}

	w.Header().Set("Content-Type", "text/html")

	if err := pageTemplate.Execute(w, data); err != nil {
		log.Printf("Template error: %v", err)
	}
}

func downloadBinary(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path[len("/download/"):])

	// Security: only allow known file prefixes and the LICENSE file.
	if filepath.Dir(filename) != "." {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(filename, binaryPrefix) && filename != licenseFile {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(binaryDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", "attachment; filename="+filename)

	if filename == licenseFile {
		w.Header().Set("Content-Type", "text/plain")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	http.ServeFile(w, r, filePath)

	log.Printf("Downloaded: %s from %s", filename, r.RemoteAddr)
}
