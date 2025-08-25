package scheduler

import (
    "database/sql"
    "log"
    "time"

    "student-to-do-list/backend/models"
    "github.com/go-co-op/gocron"
)

var scheduler *gocron.Scheduler

func InitScheduler() {
    scheduler = gocron.NewScheduler(time.Local)
    scheduler.SingletonModeAll()
    scheduler.StartAsync()
}

func ScheduleTaskAt(taskID int, taskName, scheduledTime string) {
    taskTime, err := time.ParseInLocation("2006-01-02 15:04", scheduledTime, time.Local)
    if err != nil {
        log.Printf("❌ Invalid task time for Task ID %d: %v", taskID, err)
        return
    }

    duration := time.Until(taskTime)
    if duration <= 0 {
        log.Printf("❌ Task time already passed for Task ID %d", taskID)
        return
    }

    log.Printf("🕒 Task ID %d scheduled to run after %v (at %s)", taskID, duration, scheduledTime)

    time.AfterFunc(duration, func() {
        start := time.Now()
        log.Printf("🚀 DO TASK: %s (Task ID: %d)", taskName, taskID)

        if err := models.UpdateStartTime(taskID, start); err != nil {
            log.Printf("⚠️ Failed to update start time: %v", err)
        }

        // 🔎 read duration_sec for this task
        var tmp sql.NullInt64
        durSec := int64(2)
        err := models.DB().QueryRow("SELECT COALESCE(duration_sec, 2) FROM tasks WHERE id = ?", taskID).
            Scan(&tmp)
        if err != nil {
            log.Printf("⚠️ Could not read duration for task %d (default 2s): %v", taskID, err)
        } else if tmp.Valid && tmp.Int64 > 0 {
            durSec = tmp.Int64
        }

        // simulate the actual runtime
        time.Sleep(time.Duration(durSec) * time.Second)

        end := time.Now()
        if err := models.UpdateEndTime(taskID, end); err != nil {
            log.Printf("⚠️ Failed to update end time: %v", err)
        }
        if err := models.MarkDone(taskID); err != nil {
            log.Printf("⚠️ Failed to mark task %d done: %v", taskID, err)
        }

        log.Printf("✅ Task %d completed at %v (duration %ds)", taskID, end.Format("2006-01-02 15:04:05"), durSec)
    })
}