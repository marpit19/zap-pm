package downloader

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"
	"time"
)

const (
	progressWidth = 40
	refreshRate   = 100 * time.Millisecond
)

// ProgressBar represents a progress bar for tracking downloads
type ProgressBar struct {
	current int64
	total   int64
	started time.Time
	active  bool
}

// NewProgressBar creates a new progress bar
func NewProgressBar(total int64) *ProgressBar {
	bar := &ProgressBar{
		total:   total,
		started: time.Now(),
		active:  true,
	}
	go bar.refresher()
	return bar
}

// NewProxyReader creates an io.Reader that updates progress
func (p *ProgressBar) NewProxyReader(reader io.Reader) io.Reader {
	return &ProgressReader{
		reader: reader,
		bar:    p,
	}
}

// Add increases the current progress
func (p *ProgressBar) Add(n int64) {
	atomic.AddInt64(&p.current, n)
}

// Current returns the current progress
func (p *ProgressBar) Current() int64 {
	return atomic.LoadInt64(&p.current)
}

// Percentage returns the current progress percentage
func (p *ProgressBar) Percentage() float64 {
	if p.total <= 0 {
		return 0
	}
	return float64(p.Current()) / float64(p.total) * 100
}

// Speed returns the current download speed in bytes per second
func (p *ProgressBar) Speed() float64 {
	duration := time.Since(p.started).Seconds()
	if duration <= 0 {
		return 0
	}
	return float64(p.Current()) / duration
}

// Finish marks the progress bar as complete
func (p *ProgressBar) Finish() {
	p.active = false
	p.render()
	fmt.Println() // Add newline after finishing
}

// refresher updates the progress bar display periodically
func (p *ProgressBar) refresher() {
	ticker := time.NewTicker(refreshRate)
	defer ticker.Stop()

	for range ticker.C {
		if !p.active {
			return
		}
		p.render()
	}
}

// render draws the progress bar
func (p *ProgressBar) render() {
	// Calculate progress bar width
	percentage := p.Percentage()
	completed := int(percentage / 100 * progressWidth)

	// Build progress bar string
	bar := strings.Builder{}
	bar.WriteString("\r[")
	for i := 0; i < progressWidth; i++ {
		if i < completed {
			bar.WriteString("=")
		} else if i == completed {
			bar.WriteString(">")
		} else {
			bar.WriteString(" ")
		}
	}
	bar.WriteString("]")

	// Add percentage and speed
	speed := p.Speed()
	var speedStr string
	switch {
	case speed > 1024*1024:
		speedStr = fmt.Sprintf("%.2f MB/s", speed/1024/1024)
	case speed > 1024:
		speedStr = fmt.Sprintf("%.2f KB/s", speed/1024)
	default:
		speedStr = fmt.Sprintf("%.0f B/s", speed)
	}

	fmt.Printf("%s %.1f%% %s", bar.String(), percentage, speedStr)
}

// ProgressReader wraps an io.Reader to track progress
type ProgressReader struct {
	reader io.Reader
	bar    *ProgressBar
}

func (r *ProgressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.bar.Add(int64(n))
	}
	if err == io.EOF {
		r.bar.Finish()
	}
	return n, err
}
