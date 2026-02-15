package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/pulse-downloader/pulse/internal/config"
	"github.com/pulse-downloader/pulse/internal/download"
	"github.com/pulse-downloader/pulse/internal/download/types"
	"github.com/pulse-downloader/pulse/internal/messages"
	"github.com/pulse-downloader/pulse/internal/utils"
	"github.com/spf13/cobra"
)

var (
	serverHost      string
	serverPort      int
	staticDir       string
	headlessVerbose bool
)

// serverCmd converts Pulse into a headless backend server
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run Pulse as a headless HTTP server",
	Long:  `Starts the Pulse HTTP server without the terminal UI. This is useful for deployments on VPS or running as a background service.`,
	Run: func(cmd *cobra.Command, args []string) {
		startHeadlessServer()
	},
}

func init() {
	serverCmd.Flags().StringVar(&serverHost, "host", "0.0.0.0", "Host interface to bind to")
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Port to listen on")
	serverCmd.Flags().StringVar(&staticDir, "static", "", "Directory to serve static files from (e.g. React build)")
	serverCmd.Flags().BoolVarP(&headlessVerbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.AddCommand(serverCmd)
}

func startHeadlessServer() {
	if headlessVerbose {
		os.Setenv("PULSE_VERBOSE", "true")
	}
	// Setup logging
	utils.Debug("Starting Pulse Headless Server on %s:%d", serverHost, serverPort)
	if staticDir != "" {
		absStatic, _ := filepath.Abs(staticDir)
		utils.Debug("Serving static files from: %s", absStatic)
	}

	// Load settings
	settings, err := config.LoadSettings()
	if err != nil {
		fmt.Printf("Warning: Failed to load settings: %v\n", err)
	}

	// Create progress channel
	progressChan := make(chan tea.Msg, 100)

	// Initialize WorkerPool
	pool := download.NewWorkerPool(progressChan, settings.General.MaxConcurrentDownloads)

	// Start progress consumer
	go consumeProgress(progressChan)

	// Create listener
	addr := fmt.Sprintf("%s:%d", serverHost, serverPort)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Printf("Error listening on %s: %v\n", addr, err)
		os.Exit(1)
	}

	// Create Dispatcher
	dispatcher := func(url, path, filename, quality string) {
		// Default path if empty
		if path == "" {
			path = settings.General.DefaultDownloadDir
			if path == "" {
				path, _ = os.Getwd()
			}
		}

		// Generate ID
		id := uuid.New().String()

		// Construct config
		// Note: We don't have the TUI's duplicate checking here easily without keeping state.
		// For a simple backend, we might assume the user (React app) handles this or we rely on
		// filename auto-renaming in the downloader.

		// For filename, if empty, downloader detects it.
		// If exists, downloader uniqueFilePath handles it.

		cfg := types.DownloadConfig{
			URL:        url,
			OutputPath: path,
			ID:         id,
			Filename:   filename,
			Quality:    quality,
			Verbose:    headlessVerbose,
			ProgressCh: progressChan,
			State:      types.NewProgressState(id, 0),
			Runtime: &types.RuntimeConfig{
				MaxConnectionsPerHost: settings.Connections.MaxConnectionsPerHost,
				MaxGlobalConnections:  settings.Connections.MaxGlobalConnections,
				UserAgent:             settings.Connections.UserAgent,
			},
		}

		utils.Debug("Dispatching download: %s -> %s", url, path)
		pool.Add(cfg)
	}

	// Start HTTP Server
	go startHTTPServer(ln, serverPort, dispatcher, staticDir)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Pulse Server running on %s\nPress Ctrl+C to stop.\n", addr)
	<-sigChan
	fmt.Println("\nShutting down...")
	pool.GracefulShutdown()
	fmt.Println("Goodbye!")
}

// consumeProgress reads messages from the worker pool.
// In TUI mode, BubbleTea handles this. In Headless, we just log.
// A more advanced version would maintain state for API polling.
func consumeProgress(ch <-chan tea.Msg) {
	for msg := range ch {
		switch m := msg.(type) {
		case messages.DownloadStartedMsg:
			utils.Debug("STARTED: %s (%s)", m.Filename, m.URL)
		case messages.DownloadCompleteMsg:
			utils.Debug("COMPLETED: %s (Time: %s)", "Download", m.Elapsed)
		case messages.DownloadErrorMsg:
			utils.Debug("ERROR: %s: %v", m.DownloadID, m.Err)
		case messages.ProgressMsg:
			// Verbose logging only
			// utils.Debug("Progress: %s - %.2f%%", m.DownloadID, m.Percentage*100)
		case progress.FrameMsg:
			// Ignore animation frames
		default:
			// utils.Debug("Unknown message: %T", msg)
		}
	}
}
