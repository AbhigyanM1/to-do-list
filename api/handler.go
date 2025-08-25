package api

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"student-to-do-list/backend/config"
	"student-to-do-list/backend/models"
	"student-to-do-list/backend/scheduler"

	"github.com/gin-gonic/gin"
)

/* ------------------------- Add / List / Mutations ------------------------- */

// POST /tasks
func AddTask(c *gin.Context) {
	var newTask models.Task
	if err := c.ShouldBindJSON(&newTask); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Parse scheduled_time
	sched, err := time.Parse("2006-01-02T15:04", newTask.ScheduledTime)
	if err != nil {
		sched, err = time.Parse("2006-01-02 15:04", newTask.ScheduledTime)
		if err != nil {
			log.Println("Invalid task time:", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task time format"})
			return
		}
	}
	newTask.ScheduledTime = sched.Format("2006-01-02 15:04")

	// Set arrival now
	newTask.ArrivalTime = time.Now().Format("2006-01-02 15:04:05")

	// Default duration if missing/invalid
	if newTask.DurationSec <= 0 {
		newTask.DurationSec = 2
	}

	// ✅ 4 placeholders AND 4 values
	stmt, err := config.DB.Prepare(`
        INSERT INTO tasks (name, scheduled_time, arrival_time, duration_sec)
        VALUES (?, ?, ?, ?)
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare statement"})
		return
	}

	// ✅ pass duration as the 4th argument
	res, err := stmt.Exec(newTask.Name, newTask.ScheduledTime, newTask.ArrivalTime, newTask.DurationSec)
	if err != nil {
		log.Printf("Insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert task", "detail": err.Error()})
		return
	}

	lastID, _ := res.LastInsertId()
	newTask.ID = int(lastID)

	// schedule
	scheduler.ScheduleTaskAt(newTask.ID, newTask.Name, newTask.ScheduledTime)

	c.JSON(http.StatusOK, newTask)
}

// GET /tasks
func GetTasks(c *gin.Context) {
	var tasks []models.Task
	rows, err := config.DB.Query(`
		SELECT id, name, scheduled_time, done, arrival_time, start_time, end_time
		FROM tasks
		ORDER BY id ASC`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var t models.Task
		if err := rows.Scan(
			&t.ID, &t.Name, &t.ScheduledTime, &t.Done, &t.ArrivalTime, &t.StartTime, &t.EndTime,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan task"})
			return
		}
		tasks = append(tasks, t)
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// PATCH /tasks/:id
func MarkTaskDone(c *gin.Context) {
	id := c.Param("id")
	stmt, err := config.DB.Prepare(`UPDATE tasks SET done = 1 WHERE id = ?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare update"})
		return
	}
	res, err := stmt.Exec(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task marked as done"})
}

// DELETE /tasks/:id
func DeleteTask(c *gin.Context) {
	id := c.Param("id")
	stmt, err := config.DB.Prepare(`DELETE FROM tasks WHERE id = ?`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare delete"})
		return
	}
	res, err := stmt.Exec(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

/* ------------------------------- /metrics -------------------------------- */

// single point for chart series
type metricItem struct {
	TaskID         int     `json:"task_id"`
	WaitingTime    float64 `json:"waiting_time"`
	TurnaroundTime float64 `json:"turnaround_time"`
}

type metricResponse struct {
	Algo       string       `json:"algo"`
	Metrics    []metricItem `json:"metrics"`
	AvgWaiting float64      `json:"avg_waiting"`
	AvgTAT     float64      `json:"avg_tat"`
	Throughput float64      `json:"throughput"`
}

// internal record for simulation (arrival from scheduled_time + burst from end-start)
type simRec struct {
	id      int
	arrival time.Time
	burst   float64 // seconds
}

// GET /metrics?algo=fcfs|ljf|lifo|all
func GetMetrics(c *gin.Context) {
	algo := strings.ToLower(strings.TrimSpace(c.Query("algo")))
	if algo == "" {
		algo = "fcfs"
	}

	// We take scheduled_time as the job "arrival" for simulation, and compute burst = end - start.
	// Skip rows that don't have start & end yet, because burst would be unknown.
	rows, err := config.DB.Query(`
		SELECT id, scheduled_time, start_time, end_time
		FROM tasks
		WHERE scheduled_time IS NOT NULL
		  AND start_time     IS NOT NULL
		  AND end_time       IS NOT NULL
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics"})
		return
	}
	defer rows.Close()

	const (
		layoutSched = "2006-01-02 15:04"
		layoutSE    = "2006-01-02 15:04:05"
	)

	var base []simRec
	for rows.Next() {
		var id int
		var schedStr, startStr, endStr string
		if err := rows.Scan(&id, &schedStr, &startStr, &endStr); err != nil {
			continue
		}

		arr, err1 := time.ParseInLocation(layoutSched, schedStr, time.Local)
		st, err2 := time.ParseInLocation(layoutSE, startStr, time.Local)
		et, err3 := time.ParseInLocation(layoutSE, endStr, time.Local)
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}

		burst := et.Sub(st).Seconds()
		if burst <= 0 {
			burst = 1 // guard against clock skew or missing data
		}

		base = append(base, simRec{id: id, arrival: arr, burst: burst})
	}

	// No data yet — preserve your original shape(s)
	if len(base) == 0 {
		if algo == "all" {
			c.JSON(http.StatusOK, gin.H{"results": []metricResponse{}})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"algo":        "fcfs",
			"metrics":     []metricItem{},
			"avg_waiting": 0,
			"avg_tat":     0,
			"throughput":  0,
		})
		return
	}

	makeOne := func(policy string) metricResponse {
		// Copy inputs so each policy runs independently
		tasks := make([]simRec, len(base))
		copy(tasks, base)

		// Always admit by arrival order
		sort.Slice(tasks, func(i, j int) bool { return tasks[i].arrival.Before(tasks[j].arrival) })

		type pending struct{ simRec }
		var (
			nextIdx      = 0                // next to admit by arrival
			now          = tasks[0].arrival // simulation clock
			ready        []pending          // ready pool
			done         []metricItem       // output
			sumWT        float64
			sumTAT       float64
			firstArr     = tasks[0].arrival
			lastComplete = tasks[0].arrival
		)

		admit := func() {
			for nextIdx < len(tasks) && !tasks[nextIdx].arrival.After(now) {
				ready = append(ready, pending{tasks[nextIdx]})
				nextIdx++
			}
		}

		for len(done) < len(tasks) {
			admit()
			if len(ready) == 0 {
				// jump to next arrival if idle
				if nextIdx < len(tasks) {
					now = tasks[nextIdx].arrival
					admit()
				}
			}
			if len(ready) == 0 {
				break
			}

			// choose next among READY according to policy
			choose := 0
			switch policy {
			case "fcfs":
				choose = 0 // arrival order
			case "ljf":
				for i := 1; i < len(ready); i++ {
					if ready[i].burst > ready[choose].burst {
						choose = i
					}
				}
			case "lifo":
				// most recent arrival in ready
				latest := ready[0].arrival
				choose = 0
				for i := 1; i < len(ready); i++ {
					if ready[i].arrival.After(latest) {
						latest = ready[i].arrival
						choose = i
					}
				}
			default:
				choose = 0
			}

			job := ready[choose].simRec
			ready = append(ready[:choose], ready[choose+1:]...)

			if job.arrival.After(now) {
				now = job.arrival
			}

			wt := now.Sub(job.arrival).Seconds()
			comp := now.Add(time.Duration(job.burst * float64(time.Second)))
			tat := comp.Sub(job.arrival).Seconds()

			sumWT += wt
			sumTAT += tat
			done = append(done, metricItem{
				TaskID:         job.id,
				WaitingTime:    wt,
				TurnaroundTime: tat,
			})

			now = comp
			lastComplete = comp
		}

		avgWT := sumWT / float64(len(done))
		avgTAT := sumTAT / float64(len(done))
		throughput := 0.0
		totalWall := lastComplete.Sub(firstArr).Seconds()
		if totalWall > 0 {
			throughput = float64(len(done)) / totalWall
		}

		return metricResponse{
			Algo:       policy,
			Metrics:    done,
			AvgWaiting: avgWT,
			AvgTAT:     avgTAT,
			Throughput: throughput,
		}
	}

	switch algo {
	case "all":
		c.JSON(http.StatusOK, gin.H{
			"results": []metricResponse{
				makeOne("fcfs"),
				makeOne("ljf"),
				makeOne("lifo"),
			},
		})
	case "fcfs", "ljf", "lifo":
		r := makeOne(algo)
		c.JSON(http.StatusOK, gin.H{
			"algo":        r.Algo,
			"metrics":     r.Metrics,
			"avg_waiting": r.AvgWaiting,
			"avg_tat":     r.AvgTAT,
			"throughput":  r.Throughput,
		})
	default:
		r := makeOne("fcfs")
		c.JSON(http.StatusOK, gin.H{
			"algo":        r.Algo,
			"metrics":     r.Metrics,
			"avg_waiting": r.AvgWaiting,
			"avg_tat":     r.AvgTAT,
			"throughput":  r.Throughput,
		})
	}
}
