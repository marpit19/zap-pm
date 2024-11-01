package downloader

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProgressBar(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		chunks   []int64
		wantErr  bool
		validate func(*testing.T, *ProgressBar)
	}{
		{
			name:   "basic progress",
			total:  100,
			chunks: []int64{25, 25, 25, 25},
			validate: func(t *testing.T, bar *ProgressBar) {
				assert.Equal(t, int64(100), bar.Current())
				assert.Equal(t, float64(100), bar.Percentage())
			},
		},
		{
			name:   "uneven chunks",
			total:  100,
			chunks: []int64{10, 30, 45, 15},
			validate: func(t *testing.T, bar *ProgressBar) {
				assert.Equal(t, int64(100), bar.Current())
				assert.Equal(t, float64(100), bar.Percentage())
			},
		},
		{
			name:   "zero total",
			total:  0,
			chunks: []int64{10, 20},
			validate: func(t *testing.T, bar *ProgressBar) {
				assert.Equal(t, int64(30), bar.Current())
				assert.Equal(t, float64(0), bar.Percentage())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewProgressBar(tt.total)

			for _, chunk := range tt.chunks {
				bar.Add(chunk)
				time.Sleep(10 * time.Millisecond) // Allow for refresher goroutine
			}

			if tt.validate != nil {
				tt.validate(t, bar)
			}

			bar.Finish()
		})
	}
}

func TestProgressReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		bufSize int
		wantErr bool
	}{
		{
			name:    "small buffer",
			input:   "Hello, World!",
			bufSize: 2,
			wantErr: false,
		},
		{
			name:    "exact buffer",
			input:   "Hello",
			bufSize: 5,
			wantErr: false,
		},
		{
			name:    "large buffer",
			input:   "Test",
			bufSize: 10,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a progress bar
			bar := NewProgressBar(int64(len(tt.input)))

			// Create a reader with our test input
			reader := strings.NewReader(tt.input)
			progressReader := bar.NewProxyReader(reader)

			// Read the content in chunks
			buf := make([]byte, tt.bufSize)
			output := &bytes.Buffer{}

			for {
				n, err := progressReader.Read(buf)
				if err == io.EOF {
					break
				}
				if err != nil {
					if !tt.wantErr {
						t.Errorf("unexpected error: %v", err)
					}
					return
				}
				output.Write(buf[:n])
			}

			// Verify the results
			assert.Equal(t, tt.input, output.String())
			assert.Equal(t, int64(len(tt.input)), bar.Current())
		})
	}
}

func TestSpeedCalculation(t *testing.T) {
	bar := NewProgressBar(1000)

	// Add some progress
	bar.Add(500)
	time.Sleep(100 * time.Millisecond)

	speed := bar.Speed()
	assert.True(t, speed > 0, "Speed should be greater than 0")
	assert.True(t, speed < float64(500*10), "Speed should be reasonable")
}

func TestProgressBarFormatting(t *testing.T) {
	tests := []struct {
		name          string
		totalBytes    int64
		bytesRead     int64
		expectedWidth int
	}{
		{
			name:          "empty bar",
			totalBytes:    100,
			bytesRead:     0,
			expectedWidth: progressWidth + 2, // +2 for brackets
		},
		{
			name:          "half full bar",
			totalBytes:    100,
			bytesRead:     50,
			expectedWidth: progressWidth + 2,
		},
		{
			name:          "full bar",
			totalBytes:    100,
			bytesRead:     100,
			expectedWidth: progressWidth + 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bar := NewProgressBar(tt.totalBytes)
			if tt.bytesRead > 0 {
				bar.Add(tt.bytesRead)
			}

			// Let the bar render
			time.Sleep(refreshRate * 2)

			// The actual verification of the output format is tricky because
			// it's being printed directly to stdout. In a real application,
			// we might want to modify the progress bar to accept an io.Writer
			// for testing purposes.

			assert.Equal(t, tt.bytesRead, bar.Current())
			assert.Equal(t, float64(tt.bytesRead)/float64(tt.totalBytes)*100, bar.Percentage())
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	bar := NewProgressBar(1000)

	// Simulate concurrent updates
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				bar.Add(10)
				time.Sleep(10 * time.Millisecond)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	assert.Equal(t, int64(1000), bar.Current())
}
