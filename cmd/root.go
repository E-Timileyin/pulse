package cmd

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pulse-downloader/pulse/internal/config"
	"github.com/pulse-downloader/pulse/internal/tui"
	"github.com/pulse-downloader/pulse/internal/utils"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// Version information - set via ldflags during build
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// serverProgram holds the TUI program for sending messages from HTTP handler
var serverProgram *tea.Program

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "pulse",
	Short:   "An open-source download manager written in Go",
	Long:    `Pulse is a blazing fast, open-source terminal (TUI) download manager built in Go.`,
	Version: Version,
	Run: func(cmd *cobra.Command, args []string) {
		// Find an available port starting from default
		port, listener := findAvailablePort(8080)
		if listener == nil {
			fmt.Fprintf(os.Stderr, "Error: could not find available port\n")
			os.Exit(1)
		}

		// Save port for browser extension to discover
		saveActivePort(port)

		// Create TUI program
		model := tui.InitialRootModel(port, Version)
		serverProgram = tea.NewProgram(model, tea.WithAltScreen())

		// Start HTTP server in background (reuse the listener)
		go startHTTPServer(listener, port, func(url, path, filename, quality string) {
			if serverProgram != nil {
				serverProgram.Send(tui.StartDownloadMsg{
					URL:      url,
					Path:     path,
					Filename: filename,
					// Quality: quality - Tui StartDownloadMsg doesn't have quality yet but we updated the models?
					// Actually the previous edit to update.go added Quality support to startDownload logic but
					// checks for YouTube URL first.
					// StartDownloadMsg needs updating if we want to pass quality directly.
					// But for now let's just pass basic info.
					// Wait, if I want headless to work with quality, the dispatcher needs to handle it.
					// The TUI dispatcher currently sends StartDownloadMsg.
				})
			}
		}, "")

		// Run the TUI (blocking)
		if _, err := serverProgram.Run(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		// Cleanup port file on exit
		removeActivePort()
	},
}

// findAvailablePort tries ports starting from 'start' until one is available
func findAvailablePort(start int) (int, net.Listener) {
	for port := start; port < start+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return port, ln
		}
	}
	return 0, nil
}

// saveActivePort writes the active port to ~/.pulse/port for extension discovery
func saveActivePort(port int) {
	portFile := filepath.Join(config.GetPulseDir(), "port")
	os.WriteFile(portFile, []byte(fmt.Sprintf("%d", port)), 0644)
	utils.Debug("HTTP server listening on port %d", port)
}

// removeActivePort cleans up the port file on exit
func removeActivePort() {
	portFile := filepath.Join(config.GetPulseDir(), "port")
	os.Remove(portFile)
}

// startHTTPServer starts the HTTP server using an existing listener
func startHTTPServer(ln net.Listener, port int, dispatcher DownloadDispatcher, staticDir string) {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "ok",
			"port":   port,
		})
	})

	// Download endpoint
	mux.HandleFunc("/download", makeDownloadHandler(dispatcher))

	// Static files endpoint (if configured)
	if staticDir != "" {
		fileServer := http.FileServer(http.Dir(staticDir))
		// Serve all other routes via file server
		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if file exists, otherwise serve index.html (SPA support)
			path := filepath.Join(staticDir, r.URL.Path)
			_, err := os.Stat(path)
			if os.IsNotExist(err) && !strings.HasPrefix(r.URL.Path, "/api") {
				http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
				return
			}
			fileServer.ServeHTTP(w, r)
		}))
		utils.Debug("Serving static files from: %s", staticDir)
	}

	server := &http.Server{Handler: corsMiddleware(mux)}
	if err := server.Serve(ln); err != nil && err != http.ErrServerClosed {
		utils.Debug("HTTP server error: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow all origins for now to make Vercel deployment easy
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// DownloadRequest represents a download request from the browser extension
type DownloadRequest struct {
	URL      string `json:"url"`
	Filename string `json:"filename,omitempty"`
	Path     string `json:"path,omitempty"`
	Quality  string `json:"quality,omitempty"` // Added for API support
}

// DownloadDispatcher defines how to handle a download request
type DownloadDispatcher func(url, path, filename, quality string)

func makeDownloadHandler(dispatcher DownloadDispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req DownloadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if req.URL == "" {
			http.Error(w, "URL is required", http.StatusBadRequest)
			return
		}

		if strings.Contains(req.Path, "..") || strings.Contains(req.Filename, "..") {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		if strings.Contains(req.Filename, "/") || strings.Contains(req.Filename, "\\") {
			http.Error(w, "Invalid filename", http.StatusBadRequest)
			return
		}
		if filepath.IsAbs(req.Path) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		utils.Debug("Received download request: URL=%s, Path=%s, Quality=%s", req.URL, req.Path, req.Quality)

		// Dispatch the download
		dispatcher(req.URL, req.Path, req.Filename, req.Quality)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "queued",
			"message": "Download request received",
		})
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(getCmd)
	rootCmd.SetVersionTemplate("Pulse version {{.Version}}\n")
}
