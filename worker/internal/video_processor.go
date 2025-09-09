package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VideoProcessor handles all video processing operations
type VideoProcessor struct {
	config VideoProcessingConfig
}

// VideoProcessingConfig defines configuration for video processing
type VideoProcessingConfig struct {
	MaxDuration   int    // 30 seconds maximum
	TargetWidth   int    // 1280 (for 720p 16:9)
	TargetHeight  int    // 720
	WatermarkPath string // "/app/assets/watermark.png" 
	IntroPath     string // "/app/assets/intro.mp4"
	OutroPath     string // "/app/assets/outro.mp4"
	TempDir       string // "/tmp"
	VideoCodec    string // "libx264"
	VideoQuality  string // "23" (CRF value for optimal quality)
	FFmpegPath    string // "ffmpeg"
}

// NewVideoProcessor creates a new video processor with default configuration
func NewVideoProcessor() *VideoProcessor {
	return &VideoProcessor{
		config: VideoProcessingConfig{
			MaxDuration:   30,                          // Trim to 30 seconds
			TargetWidth:   1280,                        // 720p 16:9 resolution
			TargetHeight:  720,
			WatermarkPath: "/app/assets/watermark.png", // ANB watermark
			IntroPath:     "/app/assets/intro.mp4",     // Opening bumper
			OutroPath:     "/app/assets/outro.mp4",     // Closing bumper
			TempDir:       "/tmp",
			VideoCodec:    "libx264",                   // H.264 for compatibility
			VideoQuality:  "23",                        // CRF 23 = high quality
			FFmpegPath:    "ffmpeg",                    // Assumes ffmpeg is in PATH
		},
	}
}

// ProcessVideoByS3Key applies all required transformations using no cropping
func (vp *VideoProcessor) ProcessVideoByS3Key(inputData []byte, s3Key string) ([]byte, error) {
	log.Printf("Starting video processing for S3 key: %s with no cropping", s3Key)
	log.Printf("Requirements: ≤30s, 1280x720, 16:9, no audio, ANB watermark, ANB bumpers, no content cropping")
	
	// Generate unique filename from S3 key for temp files
	safeFilename := strings.ReplaceAll(strings.ReplaceAll(s3Key, "/", "_"), ".", "_")
	
	// 1. Create temporary input file
	inputFile, err := vp.createTempFile(inputData, fmt.Sprintf("input_%s.mp4", safeFilename))
	if err != nil {
		return nil, fmt.Errorf("failed to create input temp file: %w", err)
	}
	defer vp.cleanupFile(inputFile)

	// 2. Create temporary output file  
	outputFile := filepath.Join(vp.config.TempDir, fmt.Sprintf("processed_%s.mp4", safeFilename))
	defer vp.cleanupFile(outputFile)

	// 3. Validate required assets before processing
	if err := vp.validateRequiredAssets(); err != nil {
		log.Printf("Warning: %v - continuing without watermark", err)
		// Continue without watermark rather than failing completely
	}

	// 4. Execute FFmpeg processing
	if err := vp.executeFFmpegCommand(inputFile, outputFile); err != nil {
		return nil, fmt.Errorf("ffmpeg processing failed: %w", err)
	}

	// 5. Add bumpers (intro/outro) if available
	finalOutputFile := outputFile
	if vp.bumpersExist() {
		log.Printf("Bumpers found, adding intro and outro to video")
		finalOutputFile = filepath.Join(vp.config.TempDir, fmt.Sprintf("final_%s.mp4", safeFilename))
		defer vp.cleanupFile(finalOutputFile)
		
		if err := vp.addBumpers(outputFile, finalOutputFile); err != nil {
			log.Printf("Warning: Failed to add bumpers: %v - using video without bumpers", err)
			finalOutputFile = outputFile
		} else {
			log.Printf("Successfully added ANB bumpers (intro + video + outro)")
		}
	}

	// 6. Read final processed video data
	processedData, err := os.ReadFile(finalOutputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read processed file: %w", err)
	}

	log.Printf("Video processing completed for S3 key: %s. Original: %d bytes, Processed: %d bytes", 
		s3Key, len(inputData), len(processedData))
	
	return processedData, nil
}

// executeFFmpegCommand constructs and executes the optimized FFmpeg command
func (vp *VideoProcessor) executeFFmpegCommand(inputFile, outputFile string) error {
	// Check if watermark exists to decide on filter complexity
	hasWatermark := vp.watermarkExists()
	
	var args []string
	
	if hasWatermark {
		// Full command with watermark
		args = []string{
			"-i", inputFile,                                    // Input video
			"-i", vp.config.WatermarkPath,                     // ANB watermark
			"-t", fmt.Sprintf("%d", vp.config.MaxDuration),    // Maximum 30 seconds
			"-filter_complex", vp.buildVideoFilterWithWatermark(), // Video filter with watermark
			"-an",                                             // Remove audio completely
			"-c:v", vp.config.VideoCodec,                      // Codec H.264 
			"-crf", vp.config.VideoQuality,                    // Quality CRF 23
			"-preset", "medium",                               // Balance speed/quality
			"-pix_fmt", "yuv420p",                            // Compatible pixel format
			"-movflags", "+faststart",                        // Optimization for streaming
			"-y",                                             // Overwrite output if exists
			outputFile,
		}
	} else {
		// Command without watermark
		args = []string{
			"-i", inputFile,                                    // Input video
			"-t", fmt.Sprintf("%d", vp.config.MaxDuration),    // Maximum 30 seconds
			"-vf", vp.buildVideoFilterWithoutWatermark(),      // Video filter without watermark
			"-an",                                             // Remove audio completely
			"-c:v", vp.config.VideoCodec,                      // Codec H.264 
			"-crf", vp.config.VideoQuality,                    // Quality CRF 23
			"-preset", "medium",                               // Balance speed/quality
			"-pix_fmt", "yuv420p",                            // Compatible pixel format
			"-movflags", "+faststart",                        // Optimization for streaming
			"-y",                                             // Overwrite output if exists
			outputFile,
		}
	}

	log.Printf("Executing FFmpeg with no cropping: %s", strings.Join(args, " "))
	
	cmd := exec.Command(vp.config.FFmpegPath, args...)
	cmd.Stderr = os.Stderr // Show FFmpeg errors in logs
	
	return vp.executeWithTimeout(cmd, 5*time.Minute)
}

// buildVideoFilterWithWatermark builds the complex video filter with watermark
func (vp *VideoProcessor) buildVideoFilterWithWatermark() string {
	// NO cropping - keeps all content with black bars if necessary
	// 1. scale with force_original_aspect_ratio=decrease: Video fits INSIDE 1280x720
	// 2. pad=1280:720: Adds centered black bars to complete 1280x720
	// 3. scale watermark: Scales watermark to maximum 150x60 pixels
	// 4. overlay: Places scaled ANB watermark in top right corner
	
	return fmt.Sprintf(
		"[0:v]scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black[scaled];[1:v]scale=150:60:force_original_aspect_ratio=decrease[watermark];[scaled][watermark]overlay=main_w-overlay_w-10:10",
		vp.config.TargetWidth,  // 1280
		vp.config.TargetHeight, // 720  
		vp.config.TargetWidth,  // 1280
		vp.config.TargetHeight, // 720
	)
}

// buildVideoFilterWithoutWatermark builds video filter without watermark
func (vp *VideoProcessor) buildVideoFilterWithoutWatermark() string {
	// Same scaling and padding but without watermark overlay
	return fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black",
		vp.config.TargetWidth,  // 1280
		vp.config.TargetHeight, // 720  
		vp.config.TargetWidth,  // 1280
		vp.config.TargetHeight, // 720
	)
}

// createTempFile writes data to a temporary file
func (vp *VideoProcessor) createTempFile(data []byte, filename string) (string, error) {
	tempFile := filepath.Join(vp.config.TempDir, filename)
	
	if err := os.WriteFile(tempFile, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file %s: %w", tempFile, err)
	}
	
	return tempFile, nil
}

// executeWithTimeout executes a command with a safety timeout
func (vp *VideoProcessor) executeWithTimeout(cmd *exec.Cmd, timeout time.Duration) error {
	done := make(chan error, 1)
	
	go func() {
		done <- cmd.Run()
	}()
	
	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		return fmt.Errorf("ffmpeg command timed out after %v", timeout)
	}
}

// cleanupFile safely removes a temporary file
func (vp *VideoProcessor) cleanupFile(filepath string) {
	if err := os.Remove(filepath); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: failed to cleanup temp file %s: %v", filepath, err)
	}
}

// watermarkExists checks if watermark file exists
func (vp *VideoProcessor) watermarkExists() bool {
	_, err := os.Stat(vp.config.WatermarkPath)
	return err == nil
}

// validateRequiredAssets checks if required assets and tools are available
func (vp *VideoProcessor) validateRequiredAssets() error {
	// Check if ffmpeg is available
	if _, err := exec.LookPath(vp.config.FFmpegPath); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}
	
	// Check if watermark file exists (non-critical)
	if !vp.watermarkExists() {
		return fmt.Errorf("watermark file not found at %s", vp.config.WatermarkPath)
	}
	
	log.Printf("Video processor validation passed - FFmpeg and watermark available")
	return nil
}

// bumpersExist checks if both intro and outro bumper files exist
func (vp *VideoProcessor) bumpersExist() bool {
	introExists := vp.fileExists(vp.config.IntroPath)
	outroExists := vp.fileExists(vp.config.OutroPath)
	return introExists && outroExists
}

// fileExists checks if a file exists
func (vp *VideoProcessor) fileExists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

// addBumpers concatenates intro + processed video + outro using FFmpeg
func (vp *VideoProcessor) addBumpers(processedVideoFile, outputFile string) error {
	log.Printf("Concatenating: intro + video + outro")
	
	// Create concat list file for FFmpeg
	concatFile := filepath.Join(vp.config.TempDir, fmt.Sprintf("concat_%d.txt", time.Now().Unix()))
	defer vp.cleanupFile(concatFile)
	
	concatContent := fmt.Sprintf("file '%s'\nfile '%s'\nfile '%s'\n", 
		vp.config.IntroPath, processedVideoFile, vp.config.OutroPath)
	
	if err := os.WriteFile(concatFile, []byte(concatContent), 0644); err != nil {
		return fmt.Errorf("failed to create concat file: %w", err)
	}
	
	// Execute FFmpeg concat command
	args := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-c", "copy",  // Copy streams without re-encoding for speed
		"-y",
		outputFile,
	}
	
	log.Printf("Executing FFmpeg concatenation: %s %s", vp.config.FFmpegPath, strings.Join(args, " "))
	
	cmd := exec.Command(vp.config.FFmpegPath, args...)
	cmd.Stderr = os.Stderr
	
	return vp.executeWithTimeout(cmd, 2*time.Minute)
}
