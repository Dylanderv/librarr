package api

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/JeremiahM37/librarr/internal/config"
)

// Async conversion jobs for the /api/files/* surface.
//
// Rasterizing a thick PDF outlasts the server's 60 s write timeout, so
// conversions cannot stay request/response: the POST only validates and
// enqueues (202 + job id), a single FIFO worker converts in the background,
// and the front end polls GET .../jobs/{id} until done. Closing the tab loses
// nothing — the job keeps running and the listing shows the result.
//
// One worker, in FIFO order: conversions are CPU- and disk-heavy, and the NAS
// thanks us for the restraint. Jobs live in memory only (a restart drops
// them, and a poll then 404s); the last 200 terminal jobs are kept for
// late pollers.

// Conversion job statuses, mirrored from the telegram sidecar's vocabulary.
const (
	convertJobQueued  = "queued"
	convertJobRunning = "running"
	convertJobDone    = "done"
	convertJobError   = "error"
)

// maxConvertJobs caps the retained job history; beyond it the oldest finished
// jobs are dropped.
const maxConvertJobs = 200

// convertJob is one queued or finished conversion.
type convertJob struct {
	ID       string `json:"id"`
	Batch    bool   `json:"batch"`
	Status   string `json:"status"`
	Total    int    `json:"total"`
	Finished int    `json:"finished"`
	Current  string `json:"current,omitempty"`
	Result   any    `json:"result,omitempty"`
	Error    string `json:"error,omitempty"`

	run  func(*convertJob)
	done chan struct{}
}

// convertJobManager owns the job table and the single background worker.
type convertJobManager struct {
	mu      sync.Mutex
	cfg     *config.Config
	jobs    map[string]*convertJob
	order   []string
	queue   chan string
	started bool
}

func newConvertJobManager(cfg *config.Config) *convertJobManager {
	return &convertJobManager{
		cfg:   cfg,
		jobs:  map[string]*convertJob{},
		queue: make(chan string, maxConvertJobs),
	}
}

func newConvertJobID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "job"
	}
	return "conv-" + hex.EncodeToString(b[:])
}

// enqueue registers a job and starts the worker on first use. The run closure
// executes on the worker, outside the table lock.
func (m *convertJobManager) enqueue(batch bool, total int, run func(*convertJob)) *convertJob {
	m.mu.Lock()
	job := &convertJob{
		ID:     newConvertJobID(),
		Batch:  batch,
		Status: convertJobQueued,
		Total:  total,
		run:    run,
		done:   make(chan struct{}),
	}
	m.jobs[job.ID] = job
	m.order = append(m.order, job.ID)
	for len(m.order) > maxConvertJobs {
		oldest := m.order[0]
		m.order = m.order[1:]
		if j, ok := m.jobs[oldest]; ok && (j.Status == convertJobDone || j.Status == convertJobError) {
			delete(m.jobs, oldest)
		} else {
			// Still active: keep it and stop trimming.
			m.order = append([]string{oldest}, m.order...)
			break
		}
	}
	if !m.started {
		m.started = true
		go m.work()
	}
	out := job.snapshot()
	m.mu.Unlock()
	m.queue <- job.ID
	return out
}

// work pulls jobs strictly one at a time.
func (m *convertJobManager) work() {
	for id := range m.queue {
		m.mu.Lock()
		job, ok := m.jobs[id]
		if ok {
			job.Status = convertJobRunning
		}
		run := job.run
		m.mu.Unlock()
		if !ok {
			continue
		}
		run(job)
	}
}

// snapshot copies a job for JSON output. Callers must not hold the lock;
// run closures must report through progress/finish/fail instead.
func (j *convertJob) snapshot() *convertJob {
	return &convertJob{
		ID:       j.ID,
		Batch:    j.Batch,
		Status:   j.Status,
		Total:    j.Total,
		Finished: j.Finished,
		Current:  j.Current,
		Result:   j.Result,
		Error:    j.Error,
	}
}

// get returns a copy of a job for handlers.
func (m *convertJobManager) get(id string) (*convertJob, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	j, ok := m.jobs[id]
	if !ok {
		return nil, false
	}
	return j.snapshot(), true
}

func (m *convertJobManager) progress(id string, finished int, current string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if j, ok := m.jobs[id]; ok {
		j.Finished = finished
		j.Current = current
	}
}

func (m *convertJobManager) finish(id string, result any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if j, ok := m.jobs[id]; ok {
		j.Status = convertJobDone
		j.Result = result
		j.Finished = j.Total
		j.Current = ""
		close(j.done)
	}
}

func (m *convertJobManager) fail(id, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if j, ok := m.jobs[id]; ok {
		j.Status = convertJobError
		j.Error = message
		close(j.done)
	}
}
