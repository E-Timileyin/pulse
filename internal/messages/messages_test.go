package messages

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

// =============================================================================
// ProgressMsg Tests
// =============================================================================

func TestProgressMsg_Creation(t *testing.T) {
	msg := ProgressMsg{
		DownloadID:        "test-123",
		Downloaded:        1024 * 1024,      // 1MB
		Total:             10 * 1024 * 1024, // 10MB
		Speed:             500000,           // 500KB/s
		ActiveConnections: 4,
	}

	if msg.DownloadID != "test-123" {
		t.Errorf("Expected DownloadID 'test-123', got %s", msg.DownloadID)
	}
	if msg.Downloaded != 1024*1024 {
		t.Errorf("Expected Downloaded 1MB, got %d", msg.Downloaded)
	}
	if msg.Total != 10*1024*1024 {
		t.Errorf("Expected Total 10MB, got %d", msg.Total)
	}
	if msg.Speed != 500000 {
		t.Errorf("Expected Speed 500000, got %f", msg.Speed)
	}
	if msg.ActiveConnections != 4 {
		t.Errorf("Expected ActiveConnections 4, got %d", msg.ActiveConnections)
	}
}

func TestProgressMsg_ZeroValues(t *testing.T) {
	var msg ProgressMsg

	if msg.DownloadID != "" {
		t.Error("Zero value DownloadID should be empty")
	}
	if msg.Downloaded != 0 {
		t.Error("Zero value Downloaded should be 0")
	}
	if msg.Total != 0 {
		t.Error("Zero value Total should be 0")
	}
	if msg.Speed != 0 {
		t.Error("Zero value Speed should be 0")
	}
	if msg.ActiveConnections != 0 {
		t.Error("Zero value ActiveConnections should be 0")
	}
}

func TestProgressMsg_PercentageCalculation(t *testing.T) {
	msg := ProgressMsg{
		Downloaded: 50,
		Total:      100,
	}

	percent := float64(msg.Downloaded) / float64(msg.Total) * 100
	if percent != 50.0 {
		t.Errorf("Expected 50%%, got %f%%", percent)
	}
}

func TestProgressMsg_CompleteProgress(t *testing.T) {
	msg := ProgressMsg{
		Downloaded: 1000,
		Total:      1000,
	}

	if msg.Downloaded != msg.Total {
		t.Error("Downloaded should equal Total when complete")
	}
}

func TestProgressMsg_LargeValues(t *testing.T) {
	// Test with 100GB file
	msg := ProgressMsg{
		DownloadID:        "large-download",
		Downloaded:        50 * 1024 * 1024 * 1024,  // 50GB
		Total:             100 * 1024 * 1024 * 1024, // 100GB
		Speed:             100 * 1024 * 1024,        // 100MB/s
		ActiveConnections: 32,
	}

	if msg.Downloaded != 50*1024*1024*1024 {
		t.Error("Large Downloaded value not preserved")
	}
	if msg.Total != 100*1024*1024*1024 {
		t.Error("Large Total value not preserved")
	}
}

// =============================================================================
// DownloadCompleteMsg Tests
// =============================================================================

func TestDownloadCompleteMsg_Creation(t *testing.T) {
	elapsed := 5 * time.Minute
	msg := DownloadCompleteMsg{
		DownloadID: "complete-123",
		Filename:   "test-file.zip",
		Elapsed:    elapsed,
		Total:      1024 * 1024 * 100, // 100MB
	}

	if msg.DownloadID != "complete-123" {
		t.Errorf("Expected DownloadID 'complete-123', got %s", msg.DownloadID)
	}
	if msg.Filename != "test-file.zip" {
		t.Errorf("Expected Filename 'test-file.zip', got %s", msg.Filename)
	}
	if msg.Elapsed != elapsed {
		t.Errorf("Expected Elapsed %v, got %v", elapsed, msg.Elapsed)
	}
	if msg.Total != 100*1024*1024 {
		t.Errorf("Expected Total 100MB, got %d", msg.Total)
	}
}

func TestDownloadCompleteMsg_ZeroValues(t *testing.T) {
	var msg DownloadCompleteMsg

	if msg.DownloadID != "" {
		t.Error("Zero value DownloadID should be empty")
	}
	if msg.Filename != "" {
		t.Error("Zero value Filename should be empty")
	}
	if msg.Elapsed != 0 {
		t.Error("Zero value Elapsed should be 0")
	}
	if msg.Total != 0 {
		t.Error("Zero value Total should be 0")
	}
}

func TestDownloadCompleteMsg_SpeedCalculation(t *testing.T) {
	msg := DownloadCompleteMsg{
		Elapsed: 10 * time.Second,
		Total:   10 * 1024 * 1024, // 10MB
	}

	speedMBps := float64(msg.Total) / msg.Elapsed.Seconds() / (1024 * 1024)
	if speedMBps != 1.0 {
		t.Errorf("Expected 1 MB/s, got %f MB/s", speedMBps)
	}
}

func TestDownloadCompleteMsg_LongDuration(t *testing.T) {
	msg := DownloadCompleteMsg{
		DownloadID: "long-download",
		Filename:   "large-file.iso",
		Elapsed:    24 * time.Hour,          // 1 day
		Total:      1024 * 1024 * 1024 * 50, // 50GB
	}

	if msg.Elapsed.Hours() != 24 {
		t.Errorf("Expected 24 hours, got %f hours", msg.Elapsed.Hours())
	}
}

// =============================================================================
// DownloadErrorMsg Tests
// =============================================================================

func TestDownloadErrorMsg_Creation(t *testing.T) {
	err := errors.New("connection timeout")
	msg := DownloadErrorMsg{
		DownloadID: "error-123",
		Err:        err,
	}

	if msg.DownloadID != "error-123" {
		t.Errorf("Expected DownloadID 'error-123', got %s", msg.DownloadID)
	}
	if msg.Err != err {
		t.Error("Error should be the same instance")
	}
	if msg.Err.Error() != "connection timeout" {
		t.Errorf("Expected error message 'connection timeout', got %s", msg.Err.Error())
	}
}

func TestDownloadErrorMsg_NilError(t *testing.T) {
	msg := DownloadErrorMsg{
		DownloadID: "nil-error",
		Err:        nil,
	}

	if msg.Err != nil {
		t.Error("Nil error should remain nil")
	}
}

func TestDownloadErrorMsg_WrappedError(t *testing.T) {
	innerErr := errors.New("disk full")
	wrappedErr := fmt.Errorf("failed to write: %w", innerErr)

	msg := DownloadErrorMsg{
		DownloadID: "wrapped-error",
		Err:        wrappedErr,
	}

	if !errors.Is(msg.Err, innerErr) {
		t.Error("Should be able to unwrap to inner error")
	}
}

func TestDownloadErrorMsg_ZeroValues(t *testing.T) {
	var msg DownloadErrorMsg

	if msg.DownloadID != "" {
		t.Error("Zero value DownloadID should be empty")
	}
	if msg.Err != nil {
		t.Error("Zero value Err should be nil")
	}
}

// =============================================================================
// DownloadStartedMsg Tests
// =============================================================================

func TestDownloadStartedMsg_Creation(t *testing.T) {
	msg := DownloadStartedMsg{
		DownloadID: "started-123",
		URL:        "https://example.com/file.zip",
		Filename:   "file.zip",
		Total:      50 * 1024 * 1024, // 50MB
		DestPath:   "/home/user/Downloads/file.zip",
	}

	if msg.DownloadID != "started-123" {
		t.Errorf("Expected DownloadID 'started-123', got %s", msg.DownloadID)
	}
	if msg.URL != "https://example.com/file.zip" {
		t.Errorf("Expected URL 'https://example.com/file.zip', got %s", msg.URL)
	}
	if msg.Filename != "file.zip" {
		t.Errorf("Expected Filename 'file.zip', got %s", msg.Filename)
	}
	if msg.Total != 50*1024*1024 {
		t.Errorf("Expected Total 50MB, got %d", msg.Total)
	}
	if msg.DestPath != "/home/user/Downloads/file.zip" {
		t.Errorf("Expected DestPath '/home/user/Downloads/file.zip', got %s", msg.DestPath)
	}
}

func TestDownloadStartedMsg_ZeroValues(t *testing.T) {
	var msg DownloadStartedMsg

	if msg.DownloadID != "" {
		t.Error("Zero value DownloadID should be empty")
	}
	if msg.URL != "" {
		t.Error("Zero value URL should be empty")
	}
	if msg.Filename != "" {
		t.Error("Zero value Filename should be empty")
	}
	if msg.Total != 0 {
		t.Error("Zero value Total should be 0")
	}
	if msg.DestPath != "" {
		t.Error("Zero value DestPath should be empty")
	}
}

func TestDownloadStartedMsg_UnknownSize(t *testing.T) {
	// Sometimes Content-Length is unknown
	msg := DownloadStartedMsg{
		DownloadID: "unknown-size",
		URL:        "https://example.com/stream",
		Filename:   "stream.bin",
		Total:      0, // Unknown size
		DestPath:   "/tmp/stream.bin",
	}

	if msg.Total != 0 {
		t.Error("Unknown size should be 0")
	}
}

func TestDownloadStartedMsg_LongURL(t *testing.T) {
	longURL := "https://example.com/" + string(make([]byte, 2000))
	msg := DownloadStartedMsg{
		URL: longURL,
	}

	if len(msg.URL) != len(longURL) {
		t.Error("Long URL should be preserved")
	}
}

// =============================================================================
// DownloadPausedMsg Tests
// =============================================================================

func TestDownloadPausedMsg_Creation(t *testing.T) {
	msg := DownloadPausedMsg{
		DownloadID: "paused-123",
		Downloaded: 25 * 1024 * 1024, // 25MB paused
	}

	if msg.DownloadID != "paused-123" {
		t.Errorf("Expected DownloadID 'paused-123', got %s", msg.DownloadID)
	}
	if msg.Downloaded != 25*1024*1024 {
		t.Errorf("Expected Downloaded 25MB, got %d", msg.Downloaded)
	}
}

func TestDownloadPausedMsg_ZeroValues(t *testing.T) {
	var msg DownloadPausedMsg

	if msg.DownloadID != "" {
		t.Error("Zero value DownloadID should be empty")
	}
	if msg.Downloaded != 0 {
		t.Error("Zero value Downloaded should be 0")
	}
}

func TestDownloadPausedMsg_ImmediatePause(t *testing.T) {
	// Paused immediately after starting
	msg := DownloadPausedMsg{
		DownloadID: "immediate-pause",
		Downloaded: 0,
	}

	if msg.Downloaded != 0 {
		t.Error("Immediate pause should have 0 bytes downloaded")
	}
}

// =============================================================================
// DownloadResumedMsg Tests
// =============================================================================

func TestDownloadResumedMsg_Creation(t *testing.T) {
	msg := DownloadResumedMsg{
		DownloadID: "resumed-123",
	}

	if msg.DownloadID != "resumed-123" {
		t.Errorf("Expected DownloadID 'resumed-123', got %s", msg.DownloadID)
	}
}

func TestDownloadResumedMsg_ZeroValues(t *testing.T) {
	var msg DownloadResumedMsg

	if msg.DownloadID != "" {
		t.Error("Zero value DownloadID should be empty")
	}
}

// =============================================================================
// Message Type Assertions (for interface compatibility)
// =============================================================================

func TestMessageTypes_AreDistinct(t *testing.T) {
	// Verify all message types are distinct and can be type-switched
	messages := []interface{}{
		ProgressMsg{DownloadID: "progress"},
		DownloadCompleteMsg{DownloadID: "complete"},
		DownloadErrorMsg{DownloadID: "error"},
		DownloadStartedMsg{DownloadID: "started"},
		DownloadPausedMsg{DownloadID: "paused"},
		DownloadResumedMsg{DownloadID: "resumed"},
	}

	typeNames := make(map[string]bool)
	for _, msg := range messages {
		typeName := fmt.Sprintf("%T", msg)
		if typeNames[typeName] {
			t.Errorf("Duplicate type: %s", typeName)
		}
		typeNames[typeName] = true
	}

	if len(typeNames) != 6 {
		t.Errorf("Expected 6 distinct types, got %d", len(typeNames))
	}
}

func TestMessageTypes_TypeSwitch(t *testing.T) {
	var msg interface{} = ProgressMsg{DownloadID: "test"}

	switch m := msg.(type) {
	case ProgressMsg:
		if m.DownloadID != "test" {
			t.Error("Type switch should preserve value")
		}
	default:
		t.Error("Should match ProgressMsg")
	}
}

// =============================================================================
// Channel Communication Tests
// =============================================================================

func TestProgressMsg_ChannelCommunication(t *testing.T) {
	ch := make(chan ProgressMsg, 1)

	sent := ProgressMsg{
		DownloadID: "channel-test",
		Downloaded: 1000,
		Total:      2000,
	}

	ch <- sent
	received := <-ch

	if received != sent {
		t.Error("Message should be identical after channel send/receive")
	}
}

func TestDownloadCompleteMsg_ChannelCommunication(t *testing.T) {
	ch := make(chan DownloadCompleteMsg, 1)

	sent := DownloadCompleteMsg{
		DownloadID: "channel-complete",
		Elapsed:    5 * time.Second,
	}

	ch <- sent
	received := <-ch

	if received.DownloadID != sent.DownloadID {
		t.Error("DownloadID should match")
	}
	if received.Elapsed != sent.Elapsed {
		t.Error("Elapsed should match")
	}
}

func TestDownloadErrorMsg_ChannelCommunication(t *testing.T) {
	ch := make(chan DownloadErrorMsg, 1)

	err := errors.New("test error")
	sent := DownloadErrorMsg{
		DownloadID: "channel-error",
		Err:        err,
	}

	ch <- sent
	received := <-ch

	if received.Err.Error() != err.Error() {
		t.Error("Error should match")
	}
}

// =============================================================================
// Edge Cases and Special Characters
// =============================================================================

func TestDownloadStartedMsg_SpecialFilenames(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
	}{
		{"with spaces", "my file.zip"},
		{"unicode", "文件.zip"},
		{"special chars", "file (1).zip"},
		{"very long", string(make([]byte, 255))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := DownloadStartedMsg{
				Filename: tc.filename,
			}
			if msg.Filename != tc.filename {
				t.Errorf("Filename not preserved: %s", tc.filename)
			}
		})
	}
}

func TestDownloadStartedMsg_URLVariants(t *testing.T) {
	testCases := []struct {
		name string
		url  string
	}{
		{"http", "http://example.com/file"},
		{"https", "https://example.com/file"},
		{"with port", "https://example.com:8080/file"},
		{"with query", "https://example.com/file?key=value"},
		{"with fragment", "https://example.com/file#section"},
		{"ftp", "ftp://example.com/file"},
		{"ipv4", "http://192.168.1.1/file"},
		{"ipv6", "http://[::1]/file"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := DownloadStartedMsg{
				URL: tc.url,
			}
			if msg.URL != tc.url {
				t.Errorf("URL not preserved: %s", tc.url)
			}
		})
	}
}

// =============================================================================
// Equality and Comparison Tests
// =============================================================================

func TestProgressMsg_Equality(t *testing.T) {
	msg1 := ProgressMsg{
		DownloadID:        "equal",
		Downloaded:        100,
		Total:             200,
		Speed:             50,
		ActiveConnections: 2,
	}
	msg2 := ProgressMsg{
		DownloadID:        "equal",
		Downloaded:        100,
		Total:             200,
		Speed:             50,
		ActiveConnections: 2,
	}

	if msg1 != msg2 {
		t.Error("Identical ProgressMsg should be equal")
	}
}

func TestDownloadCompleteMsg_Equality(t *testing.T) {
	elapsed := 5 * time.Second
	msg1 := DownloadCompleteMsg{
		DownloadID: "equal",
		Filename:   "file.zip",
		Elapsed:    elapsed,
		Total:      1000,
	}
	msg2 := DownloadCompleteMsg{
		DownloadID: "equal",
		Filename:   "file.zip",
		Elapsed:    elapsed,
		Total:      1000,
	}

	if msg1 != msg2 {
		t.Error("Identical DownloadCompleteMsg should be equal")
	}
}

// Note: DownloadErrorMsg equality is tricky because error comparison
// compares pointer/interface, not value

func TestDownloadPausedMsg_Equality(t *testing.T) {
	msg1 := DownloadPausedMsg{DownloadID: "equal", Downloaded: 500}
	msg2 := DownloadPausedMsg{DownloadID: "equal", Downloaded: 500}

	if msg1 != msg2 {
		t.Error("Identical DownloadPausedMsg should be equal")
	}
}

func TestDownloadResumedMsg_Equality(t *testing.T) {
	msg1 := DownloadResumedMsg{DownloadID: "equal"}
	msg2 := DownloadResumedMsg{DownloadID: "equal"}

	if msg1 != msg2 {
		t.Error("Identical DownloadResumedMsg should be equal")
	}
}
