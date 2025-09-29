package jobs

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// VideoJob represents a video processing job
type VideoJob struct {
	ID           int64     `json:"id"`
	VideoID      int64     `json:"video_id"`
	Status       JobStatus `json:"status"`
	FilePath     string    `json:"file_path"`
	OutputPath   *string   `json:"output_path"`
	ErrorMessage *string   `json:"error_message"`
	Attempts     int       `json:"attempts"`
	MaxAttempts  int       `json:"max_attempts"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ProcessedAt  *time.Time `json:"processed_at"`
}

// JobQueue handles job queue operations
type JobQueue struct {
	db           *sql.DB
	pollInterval time.Duration
}

// NewJobQueue creates a new job queue
func NewJobQueue(db *sql.DB, pollInterval time.Duration) *JobQueue {
	if pollInterval == 0 {
		pollInterval = 5 * time.Second // Default poll interval
	}

	return &JobQueue{
		db:           db,
		pollInterval: pollInterval,
	}
}

// EnqueueJob adds a new job to the queue
func (jq *JobQueue) EnqueueJob(ctx context.Context, videoID int64, filePath string) (*VideoJob, error) {
	query := `
		INSERT INTO video_jobs (video_id, file_path, status) 
		VALUES (?, ?, ?)
	`

	result, err := jq.db.ExecContext(ctx, query, videoID, filePath, JobStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to enqueue job: %w", err)
	}

	jobID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get job ID: %w", err)
	}

	// Fetch the created job
	return jq.GetJob(ctx, jobID)
}

// DequeueJob gets the next pending job and marks it as processing
func (jq *JobQueue) DequeueJob(ctx context.Context) (*VideoJob, error) {
	tx, err := jq.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock and get next pending job
	query := `
		SELECT id, video_id, status, file_path, output_path, error_message, 
		       attempts, max_attempts, created_at, updated_at, processed_at
		FROM video_jobs 
		WHERE status = ? AND attempts < max_attempts
		ORDER BY created_at ASC 
		LIMIT 1 
		FOR UPDATE
	`

	var job VideoJob
	err = tx.QueryRowContext(ctx, query, JobStatusPending).Scan(
		&job.ID, &job.VideoID, &job.Status, &job.FilePath, &job.OutputPath,
		&job.ErrorMessage, &job.Attempts, &job.MaxAttempts, &job.CreatedAt,
		&job.UpdatedAt, &job.ProcessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to query job: %w", err)
	}

	// Update job status to processing
	updateQuery := `
		UPDATE video_jobs 
		SET status = ?, attempts = attempts + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err = tx.ExecContext(ctx, updateQuery, JobStatusProcessing, job.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update job status: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	job.Status = JobStatusProcessing
	job.Attempts++

	return &job, nil
}

// CompleteJob marks a job as completed
func (jq *JobQueue) CompleteJob(ctx context.Context, jobID int64, outputPath string) error {
	query := `
		UPDATE video_jobs 
		SET status = ?, output_path = ?, processed_at = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := jq.db.ExecContext(ctx, query, JobStatusCompleted, outputPath, jobID)
	if err != nil {
		return fmt.Errorf("failed to complete job: %w", err)
	}

	return nil
}

// FailJob marks a job as failed
func (jq *JobQueue) FailJob(ctx context.Context, jobID int64, errorMessage string) error {
	query := `
		UPDATE video_jobs 
		SET status = ?, error_message = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := jq.db.ExecContext(ctx, query, JobStatusFailed, errorMessage, jobID)
	if err != nil {
		return fmt.Errorf("failed to fail job: %w", err)
	}

	return nil
}

// RetryJob resets a failed job to pending status if it hasn't exceeded max attempts
func (jq *JobQueue) RetryJob(ctx context.Context, jobID int64) error {
	query := `
		UPDATE video_jobs 
		SET status = ?, error_message = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND attempts < max_attempts
	`

	result, err := jq.db.ExecContext(ctx, query, JobStatusPending, jobID)
	if err != nil {
		return fmt.Errorf("failed to retry job: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job not found or max attempts exceeded")
	}

	return nil
}

// GetJob retrieves a job by ID
func (jq *JobQueue) GetJob(ctx context.Context, jobID int64) (*VideoJob, error) {
	query := `
		SELECT id, video_id, status, file_path, output_path, error_message,
		       attempts, max_attempts, created_at, updated_at, processed_at
		FROM video_jobs 
		WHERE id = ?
	`

	var job VideoJob
	err := jq.db.QueryRowContext(ctx, query, jobID).Scan(
		&job.ID, &job.VideoID, &job.Status, &job.FilePath, &job.OutputPath,
		&job.ErrorMessage, &job.Attempts, &job.MaxAttempts, &job.CreatedAt,
		&job.UpdatedAt, &job.ProcessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("job not found")
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return &job, nil
}

// PollForJobs continuously polls for jobs and processes them
func (jq *JobQueue) PollForJobs(ctx context.Context, handler func(*VideoJob) error) error {
	log.Printf("Starting job polling with interval: %v", jq.pollInterval)

	ticker := time.NewTicker(jq.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Job polling stopped")
			return ctx.Err()
		case <-ticker.C:
			job, err := jq.DequeueJob(ctx)
			if err != nil {
				log.Printf("Error dequeuing job: %v", err)
				continue
			}

			if job == nil {
				// No jobs available, continue polling
				continue
			}

			log.Printf("Processing job %d for video %d", job.ID, job.VideoID)

			// Process the job
			err = handler(job)
			if err != nil {
				log.Printf("Error processing job %d: %v", job.ID, err)
				if failErr := jq.FailJob(ctx, job.ID, err.Error()); failErr != nil {
					log.Printf("Error marking job as failed: %v", failErr)
				}
			}
		}
	}
}

// GetJobsByStatus retrieves jobs by status
func (jq *JobQueue) GetJobsByStatus(ctx context.Context, status JobStatus, limit int) ([]*VideoJob, error) {
	query := `
		SELECT id, video_id, status, file_path, output_path, error_message,
		       attempts, max_attempts, created_at, updated_at, processed_at
		FROM video_jobs 
		WHERE status = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := jq.db.QueryContext(ctx, query, status, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*VideoJob
	for rows.Next() {
		var job VideoJob
		err := rows.Scan(
			&job.ID, &job.VideoID, &job.Status, &job.FilePath, &job.OutputPath,
			&job.ErrorMessage, &job.Attempts, &job.MaxAttempts, &job.CreatedAt,
			&job.UpdatedAt, &job.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// CleanupOldJobs removes completed/failed jobs older than specified duration
func (jq *JobQueue) CleanupOldJobs(ctx context.Context, olderThan time.Duration) error {
	cutoff := time.Now().Add(-olderThan)
	
	query := `
		DELETE FROM video_jobs 
		WHERE status IN (?, ?) AND updated_at < ?
	`

	result, err := jq.db.ExecContext(ctx, query, JobStatusCompleted, JobStatusFailed, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup old jobs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	log.Printf("Cleaned up %d old jobs", rowsAffected)

	return nil
}