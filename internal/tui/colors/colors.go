package colors

import "github.com/charmbracelet/lipgloss"

// === Color Palette ===
// Solar Gold / Luxury Dark
var (
	// Primary Accents
	Gold       = lipgloss.Color("#FFD700") // Core Gold
	BrightGold = lipgloss.Color("#FFE55C") // Highlight
	DarkGold   = lipgloss.Color("#B8860B") // Dimmed Gold

	// Creating depth
	RichBlack = lipgloss.Color("#0a0a0a") // Deep background
	SoftBlack = lipgloss.Color("#1a1a1a") // Panels
	Border    = lipgloss.Color("#333333") // Subtle borders

	// Neutrals
	White     = lipgloss.Color("#FFFFFF")
	LightGray = lipgloss.Color("#B0B0B0")
	Gray      = lipgloss.Color("#555555")

	// Aliases for compatibility with existing code
	NeonPurple = BrightGold
	NeonPink   = Gold
	NeonCyan   = White
	DarkGray   = SoftBlack
)

// === Semantic State Colors ===
var (
	StateError       = lipgloss.Color("#FF4444") // 🔴 Red - Error/Stopped
	StatePaused      = lipgloss.Color("#FFAA00") // 🟡 Amber - Paused/Queued
	StateDownloading = lipgloss.Color("#FFD700") // 🟡 Gold - Downloading (Active)
	StateDone        = lipgloss.Color("#00FF88") // 🟢 Green - Completed (Success)
)

// === Progress Bar Colors ===
const (
	ProgressStart = "#B8860B" // Dark Gold
	ProgressEnd   = "#FFD700" // Bright Gold
)
