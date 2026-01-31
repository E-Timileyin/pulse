package download

import (
	"fmt"
	"strings"

	"github.com/kkdai/youtube/v2"
	"github.com/pulse-downloader/pulse/internal/utils"
)

// IsYoutubeURL checks if the URL is a YouTube link
func IsYoutubeURL(url string) bool {
	return strings.Contains(url, "youtube.com/") || strings.Contains(url, "youtu.be/")
}

// ResolveYoutubeURL extracts the best direct download URL from a YouTube video
func ResolveYoutubeURL(videoURL, quality string) (string, string, error) {
	client := youtube.Client{}

	utils.Debug("Fetching YouTube video info for: %s (Quality: %s)", videoURL, quality)
	video, err := client.GetVideo(videoURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to get video info: %w", err)
	}

	var bestFormat *youtube.Format
	formats := video.Formats

	// Filter for progressive formats first (audio + video)
	// In the future we could support adaptive streaming (DASH) which requires merging,
	// but for now we stick to simple progressive downloads or pre-muxed streams.
	var options []youtube.Format
	for _, f := range formats {
		if f.AudioChannels > 0 && f.Width > 0 {
			options = append(options, f)
		}
	}

	if len(options) == 0 {
		return "", "", fmt.Errorf("no usable video formats found (progressive with audio)")
	}

	// Helper to find specific quality
	findQuality := func(q string) *youtube.Format {
		for _, f := range options {
			if strings.Contains(strings.ToLower(f.QualityLabel), strings.ToLower(q)) {
				return &f
			}
		}
		return nil
	}

	// If user requested a specific quality, try to find it
	if quality != "" {
		bestFormat = findQuality(quality)
		if bestFormat == nil {
			utils.Debug("Requested quality '%s' not found, falling back to best available", quality)
		}
	}

	// Fallback to highest bitrate if no specific quality found or requested
	if bestFormat == nil {
		for i := range options {
			if bestFormat == nil || options[i].Bitrate > bestFormat.Bitrate {
				bestFormat = &options[i]
			}
		}
	}

	utils.Debug("Selected format: %s (Quality: %s)",
		bestFormat.MimeType, bestFormat.QualityLabel)

	url, err := client.GetStreamURL(video, bestFormat)
	if err != nil {
		return "", "", fmt.Errorf("failed to get stream URL: %w", err)
	}

	// Clean title for filename
	title := sanitizeFilename(video.Title)
	if !strings.HasSuffix(strings.ToLower(title), ".mp4") {
		title += ".mp4"
	}

	return url, title, nil
}

// GetVideoQualities returns a list of available quality labels for a video
func GetVideoQualities(videoURL string) ([]string, string, error) {
	client := youtube.Client{}
	video, err := client.GetVideo(videoURL)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get video info: %w", err)
	}

	var qualities []string
	seen := make(map[string]bool)

	// Collect unique qualities from progressive formats
	for _, f := range video.Formats {
		if f.AudioChannels > 0 && f.Width > 0 && f.QualityLabel != "" {
			if !seen[f.QualityLabel] {
				qualities = append(qualities, f.QualityLabel)
				seen[f.QualityLabel] = true
			}
		}
	}

	title := sanitizeFilename(video.Title)
	return qualities, title, nil
}

func sanitizeFilename(name string) string {
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}
	return strings.TrimSpace(name)
}
