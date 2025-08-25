package models

import (
	"database/sql"
	"student-to-do-list/backend/config"
	"time"
)

type Task struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	ScheduledTime string `json:"scheduled_time"`
	Done          int    `json:"done"`
	ArrivalTime   string `json:"arrival_time"` // set when task is created
	StartTime     string `json:"start_time"`   // set when scheduler triggers
	EndTime       string `json:"end_time"`
	DurationSec   int    `json:"duration_sec"` // execution time for LJF/LIFO algorithms
}

// UpdateStartTime sets the actual start time in DB
func UpdateStartTime(taskID int, start time.Time) error {
	_, err := config.DB.Exec(
		"UPDATE tasks SET start_time = ? WHERE id = ?",
		start.Format("2006-01-02 15:04:05"),
		taskID,
	)
	return err
}

// UpdateEndTime sets the actual end time in DB
func UpdateEndTime(taskID int, end time.Time) error {
	_, err := config.DB.Exec(
		"UPDATE tasks SET end_time = ? WHERE id = ?",
		end.Format("2006-01-02 15:04:05"),
		taskID,
	)
	return err
}

// UpdateDuration updates the duration_sec column in DB
func UpdateDuration(taskID int, duration int) error {
	_, err := config.DB.Exec(
		"UPDATE tasks SET duration_sec = ? WHERE id = ?",
		duration,
		taskID,
	)
	return err
}

// ✅ NEW: mark a task as done (uses positional placeholder)
func MarkDone(taskID int) error {
    _, err := config.DB.Exec("UPDATE tasks SET done = 1 WHERE id = ?", taskID)
    return err
}

func DB() *sql.DB { return config.DB }
