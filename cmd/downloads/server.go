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
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

var (
	archiveDir   = getEnv("ARCHIVE_DIR", "/archives")
	port         = getEnv("PORT", "8080")
	pageTemplate = template.Must(template.ParseFS(templateFS, "templates/index.html"))
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type archiveFile struct {
	Name     string
	Size     float64
	OS       string
	Arch     string
	Checksum string
}

func main() {
	files, err := listBinaryFiles()
	if err != nil || len(files) == 0 {
		log.Fatal("No downloadable files found in ", archiveDir)
	}
	log.Printf("Found %d downloadable files", len(files))

	staticContent, err := fs.Sub(staticFS, "static")
	if err != nil {
		log.Fatal("Failed to load static files: ", err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))
	http.HandleFunc("/", listBinaries)
	http.HandleFunc("/download/", downloadBinary)

	log.Printf("Starting server on port %s", port)
	log.Printf("Serving files from %s", archiveDir)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

// listBinaryFiles returns all downloadable files (excludes sha256sum.txt and .sha256 sidecars).
func listBinaryFiles() ([]string, error) {
	entries, err := os.ReadDir(archiveDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, "oadp-vmdp_") &&
			!strings.HasSuffix(name, ".sha256") &&
			name != "sha256sum.txt" {
			files = append(files, filepath.Join(archiveDir, name))
		}
	}
	return files, nil
}

// readChecksum reads a SHA256 checksum from a .sha256 sidecar file or
// falls back to looking up the filename in sha256sum.txt.
func readChecksum(filePath string) string {
	// Try per-file .sha256 sidecar first (oadp-cli style)
	data, err := os.ReadFile(filePath + ".sha256")
	if err == nil {
		fields := strings.Fields(string(data))
		if len(fields) > 0 {
			return fields[0]
		}
	}

	// Fall back to sha256sum.txt
	sumFile := filepath.Join(filepath.Dir(filePath), "sha256sum.txt")
	data, err = os.ReadFile(sumFile)
	if err != nil {
		return ""
	}
	base := filepath.Base(filePath)
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == base {
			return fields[0]
		}
	}
	return ""
}

// parsePlatform extracts OS and architecture from a filename like
// oadp-vmdp_v1.0.0_linux_amd64 or oadp-vmdp_v1.0.0_windows_arm64.exe
func parsePlatform(filename string) (string, string) {
	name := strings.TrimSuffix(filename, ".exe")
	parts := strings.Split(name, "_")
	if len(parts) >= 3 {
		return parts[len(parts)-2], parts[len(parts)-1]
	}
	return "unknown", "unknown"
}

func listBinaries(w http.ResponseWriter, r *http.Request) {
	files, err := listBinaryFiles()
	if err != nil {
		http.Error(w, "Error listing files", http.StatusInternalServerError)
		return
	}

	var linuxFiles, windowsFiles []archiveFile
	for _, file := range files {
		name := filepath.Base(file)
		info, err := os.Stat(file)
		if err != nil {
			continue
		}
		size := float64(info.Size()) / (1024 * 1024)
		osName, arch := parsePlatform(name)
		checksum := readChecksum(file)
		af := archiveFile{Name: name, Size: size, OS: osName, Arch: arch, Checksum: checksum}
		switch osName {
		case "linux":
			linuxFiles = append(linuxFiles, af)
		case "windows":
			windowsFiles = append(windowsFiles, af)
		}
	}

	data := struct {
		LinuxFiles   []archiveFile
		WindowsFiles []archiveFile
	}{linuxFiles, windowsFiles}

	w.Header().Set("Content-Type", "text/html")
	if err := pageTemplate.Execute(w, data); err != nil {
		log.Printf("Template error: %v", err)
	}
}

func downloadBinary(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path[len("/download/"):])

	// Security: only allow known file prefixes and the checksum file
	if filepath.Dir(filename) != "." {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(filename, "oadp-vmdp_") && filename != "sha256sum.txt" {
		http.Error(w, "Invalid filename", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(archiveDir, filename)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	if filename == "sha256sum.txt" {
		w.Header().Set("Content-Type", "text/plain")
	} else {
		w.Header().Set("Content-Type", "application/octet-stream")
	}

	http.ServeFile(w, r, filePath)
	log.Printf("Downloaded: %s from %s", filename, r.RemoteAddr)
}
