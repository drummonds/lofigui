package main

import (
	"fmt"
	"sync"
	"time"

	"codeberg.org/hum3/lofigui"
)

// StoredFile tracks a file we've uploaded to SeaweedFS.
type StoredFile struct {
	Fid       string
	VolumeURL string
	Name      string
	Size      int
	CreatedAt time.Time
	Verified  bool
	Content   string
}

var (
	client *Client

	filesMu sync.Mutex
	files   []StoredFile

	autoMu     sync.Mutex
	autoActive bool
	autoStop   chan struct{}

	lastError   string
	lastErrorMu sync.Mutex

	renderMu sync.Mutex
)

func setLastError(err error) {
	lastErrorMu.Lock()
	defer lastErrorMu.Unlock()
	if err != nil {
		lastError = err.Error()
	} else {
		lastError = ""
	}
}

func getLastError() string {
	lastErrorMu.Lock()
	defer lastErrorMu.Unlock()
	return lastError
}

func isAutoActive() bool {
	autoMu.Lock()
	defer autoMu.Unlock()
	return autoActive
}

func createTestFile() error {
	name := fmt.Sprintf("test_%s.txt", time.Now().Format("150405"))
	content := fmt.Sprintf("Hello from SeaweedFS demo!\nCreated at: %s\n", time.Now().Format(time.RFC3339))

	assign, err := client.Assign()
	if err != nil {
		return fmt.Errorf("assign: %w", err)
	}

	if err := client.Upload(assign.URL, assign.Fid, name, []byte(content)); err != nil {
		return fmt.Errorf("upload: %w", err)
	}

	filesMu.Lock()
	files = append(files, StoredFile{
		Fid:       assign.Fid,
		VolumeURL: assign.URL,
		Name:      name,
		Size:      len(content),
		CreatedAt: time.Now(),
		Content:   content,
	})
	filesMu.Unlock()
	return nil
}

func verifyFile(index int) error {
	filesMu.Lock()
	if index < 0 || index >= len(files) {
		filesMu.Unlock()
		return fmt.Errorf("invalid file index")
	}
	f := files[index]
	filesMu.Unlock()

	data, err := client.Download(f.VolumeURL, f.Fid)
	if err != nil {
		return err
	}

	filesMu.Lock()
	files[index].Verified = true
	files[index].Content = string(data)
	filesMu.Unlock()
	return nil
}

func deleteFile(index int) error {
	filesMu.Lock()
	if index < 0 || index >= len(files) {
		filesMu.Unlock()
		return fmt.Errorf("invalid file index")
	}
	f := files[index]
	filesMu.Unlock()

	if err := client.Delete(f.VolumeURL, f.Fid); err != nil {
		return err
	}

	filesMu.Lock()
	files = append(files[:index], files[index+1:]...)
	filesMu.Unlock()
	return nil
}

func startAutoCreate() {
	autoMu.Lock()
	defer autoMu.Unlock()
	if autoActive {
		return
	}
	autoActive = true
	autoStop = make(chan struct{})
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		if err := createTestFile(); err != nil {
			setLastError(err)
		}
		for {
			select {
			case <-autoStop:
				return
			case <-ticker.C:
				if err := createTestFile(); err != nil {
					setLastError(err)
				}
			}
		}
	}()
}

func stopAutoCreate() {
	autoMu.Lock()
	defer autoMu.Unlock()
	if !autoActive {
		return
	}
	close(autoStop)
	autoActive = false
}

func renderAndCapture(fn func()) string {
	renderMu.Lock()
	defer renderMu.Unlock()
	lofigui.Reset()
	fn()
	return lofigui.Buffer()
}
