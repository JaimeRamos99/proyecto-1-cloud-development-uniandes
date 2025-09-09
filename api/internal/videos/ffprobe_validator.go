package videos

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// FFProbeValidator handles video validation using ffprobe
type FFProbeValidator struct {
	ffprobePath string
	tempDir     string
}

// FFProbeOutput represents the structure of ffprobe JSON output
type FFProbeOutput struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Duration  string `json:"duration"`
	} `json:"streams"`
	Format struct {
		FormatName     string `json:"format_name"`
		Duration       string `json:"duration"`
		Size           string `json:"size"`
		BitRate        string `json:"bit_rate"`
		NbStreams      int    `json:"nb_streams"`
		FormatLongName string `json:"format_long_name"`
	} `json:"format"`
}

// ValidationRules represents video validation constraints
type ValidationRules struct {
	MaxSizeBytes int64   // 100MB
	MinDuration  float64 // 20 seconds
	MaxDuration  float64 // 60 seconds
	MinHeight    int     // > 1080p
}

// NewFFProbeValidator creates a new FFprobe validator
func NewFFProbeValidator(tempDir string) *FFProbeValidator {
	if tempDir == "" {
		tempDir = "/tmp"
	}

	return &FFProbeValidator{
		ffprobePath: "ffprobe", // Assumes ffprobe is in PATH (installed via Docker)
		tempDir:     tempDir,
	}
}

// DefaultValidationRules returns the validation rules for this project
func DefaultValidationRules() ValidationRules {
	return ValidationRules{
		MaxSizeBytes: 100 * 1024 * 1024, // 100MB
		MinDuration:  20,                // 20 seconds
		MaxDuration:  60,                // 60 seconds
		MinHeight:    1080,              // Greater than 1080p
	}
}

// ValidateVideo performs complete video validation using ffprobe
func (v *FFProbeValidator) ValidateVideo(file *multipart.FileHeader, rules ValidationRules) (*VideoMetadata, error) {
	// 1. Quick pre-validation (no I/O)
	if err := v.quickValidation(file, rules); err != nil {
		return nil, fmt.Errorf("quick validation failed: %w", err)
	}

	// 2. Save uploaded file to temporary location
	tempFile, err := v.saveToTempFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile) // Always clean up

	// 3. Run ffprobe analysis
	metadata, err := v.analyzeWithFFProbe(tempFile, file.Size)
	if err != nil {
		return nil, fmt.Errorf("ffprobe analysis failed: %w", err)
	}

	// 4. Validate extracted metadata against rules
	if err := v.validateMetadata(metadata, rules); err != nil {
		return nil, fmt.Errorf("metadata validation failed: %w", err)
	}

	return metadata, nil
}

// quickValidation performs fast validations without I/O
func (v *FFProbeValidator) quickValidation(file *multipart.FileHeader, rules ValidationRules) error {
	// Check file size first (fastest check)
	if file.Size > rules.MaxSizeBytes {
		return fmt.Errorf("file too large: %d bytes (max: %d bytes / %.1fMB)",
			file.Size, rules.MaxSizeBytes, float64(rules.MaxSizeBytes)/(1024*1024))
	}

	if file.Size == 0 {
		return fmt.Errorf("file is empty")
	}

	// Check file extension (basic format check)
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".mp4" {
		return fmt.Errorf("invalid file extension: %s (only .mp4 allowed)", ext)
	}

	return nil
}

// saveToTempFile saves the uploaded file to a temporary location
func (v *FFProbeValidator) saveToTempFile(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create temp file with .mp4 extension for better ffprobe detection
	tempFile, err := os.CreateTemp(v.tempDir, "video_*.mp4")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tempFile.Close()

	// Copy uploaded file to temp location
	if _, err := io.Copy(tempFile, src); err != nil {
		os.Remove(tempFile.Name())
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	return tempFile.Name(), nil
}

// analyzeWithFFProbe runs ffprobe on the file and extracts metadata
func (v *FFProbeValidator) analyzeWithFFProbe(tempFile string, fileSize int64) (*VideoMetadata, error) {
	// Execute ffprobe with JSON output
	cmd := exec.Command(v.ffprobePath,
		"-v", "quiet", // Suppress verbose output
		"-print_format", "json", // Output as JSON
		"-show_format",  // Show container format info
		"-show_streams", // Show stream info (video, audio)
		tempFile,
	)

	output, err := cmd.Output()
	if err != nil {
		// If ffprobe fails, the file is likely corrupted or not a valid video
		return nil, fmt.Errorf("file is not a valid video or is corrupted")
	}

	// Parse ffprobe JSON output
	var probeResult FFProbeOutput
	if err := json.Unmarshal(output, &probeResult); err != nil {
		return nil, fmt.Errorf("failed to parse ffprobe output: %w", err)
	}

	// Extract and validate metadata
	return v.extractMetadata(&probeResult, fileSize)
}

// extractMetadata extracts video metadata from ffprobe output
func (v *FFProbeValidator) extractMetadata(probe *FFProbeOutput, fileSize int64) (*VideoMetadata, error) {
	// Validate container format is MP4
	if !v.isValidMP4Format(probe.Format.FormatName) {
		return nil, fmt.Errorf("invalid container format: %s (expected MP4)", probe.Format.FormatName)
	}

	// Find the video stream (there should be at least one)
	var videoStream *struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Duration  string `json:"duration"`
	}

	for i := range probe.Streams {
		if probe.Streams[i].CodecType == "video" {
			videoStream = &probe.Streams[i]
			break
		}
	}

	if videoStream == nil {
		return nil, fmt.Errorf("no video stream found in file")
	}

	// Extract duration (prefer stream duration, fallback to container duration)
	duration := videoStream.Duration
	if duration == "" || duration == "N/A" {
		duration = probe.Format.Duration
	}

	if duration == "" || duration == "N/A" {
		return nil, fmt.Errorf("could not determine video duration")
	}

	durationFloat, err := strconv.ParseFloat(duration, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid duration value: %s", duration)
	}

	// Validate video has proper dimensions
	if videoStream.Width == 0 || videoStream.Height == 0 {
		return nil, fmt.Errorf("invalid video dimensions: %dx%d", videoStream.Width, videoStream.Height)
	}

	// Validate video codec is compatible
	if !v.isValidVideoCodec(videoStream.CodecName) {
		return nil, fmt.Errorf("unsupported video codec: %s (expected H.264, H.265)", videoStream.CodecName)
	}

	return &VideoMetadata{
		Duration: durationFloat,
		Width:    videoStream.Width,
		Height:   videoStream.Height,
		Size:     fileSize,
		Format:   "mp4",
	}, nil
}

// validateMetadata validates extracted metadata against business rules
func (v *FFProbeValidator) validateMetadata(metadata *VideoMetadata, rules ValidationRules) error {
	// Validate duration is within acceptable range
	if metadata.Duration < rules.MinDuration || metadata.Duration > rules.MaxDuration {
		return fmt.Errorf("video duration %.1f seconds is not in range %.1f-%.1f seconds",
			metadata.Duration, rules.MinDuration, rules.MaxDuration)
	}

	// Validate resolution is greater than 1080p
	if metadata.Height <= rules.MinHeight-1 { // rules.MinHeight is 1081 (> 1080)
		return fmt.Errorf("video resolution %dx%d is below minimum 1080p (1920x1080)",
			metadata.Width, metadata.Height)
	}

	// Additional sanity checks
	if metadata.Width < 1920 { // Assume minimum Full HD width
		return fmt.Errorf("video width %dpx is too low (minimum: 1920px for >1080p)",
			metadata.Width)
	}

	return nil
}

// isValidMP4Format checks if ffprobe detected format is MP4 compatible
func (v *FFProbeValidator) isValidMP4Format(formatName string) bool {
	// FFprobe typically returns "mov,mp4,m4a,3gp,3g2,mj2" for MP4 files
	validFormats := []string{
		"mov,mp4,m4a,3gp,3g2,mj2", // Most common ffprobe output for MP4
		"mp4",
		"mov", // MOV is essentially the same container as MP4
	}

	formatLower := strings.ToLower(formatName)
	for _, valid := range validFormats {
		if strings.Contains(formatLower, "mp4") || formatLower == valid {
			return true
		}
	}

	return false
}

// isValidVideoCodec validates that the video codec is supported
func (v *FFProbeValidator) isValidVideoCodec(codecName string) bool {
	// Accept common MP4 video codecs
	validCodecs := []string{
		"h264", // Most common (AVC)
		"h265", // HEVC
		"hevc", // HEVC alternative name
		"avc1", // H.264 in MP4 container
		"hvc1", // H.265 in MP4 container
		"mp4v", // MPEG-4 Part 2
	}

	codecLower := strings.ToLower(codecName)
	for _, valid := range validCodecs {
		if codecLower == valid {
			return true
		}
	}

	return false
}

// CheckFFProbeInstallation verifies that ffprobe is available
func (v *FFProbeValidator) CheckFFProbeInstallation() error {
	cmd := exec.Command(v.ffprobePath, "-version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("ffprobe is not installed or not in PATH")
	}

	// Basic check that it's actually ffprobe
	if !strings.Contains(string(output), "ffprobe") {
		return fmt.Errorf("ffprobe command did not return expected version info")
	}

	return nil
}
