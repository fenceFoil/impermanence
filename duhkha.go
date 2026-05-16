package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/go-toast/toast"
)

const (
	PendingItemsFile = "data/pendingItems.jsonl"
	ActivityLogFile  = "data/activityLog.txt"
	ComputerName     = "VALJEAN"
	DeleteDelayDays  = 30
)

type PendingItem struct {
	Added                string  `json:"added"`
	Path                 string  `json:"path"`
	Computer             string  `json:"computer"`
	DeleteAt             string  `json:"deleteAt"`
	FailedDeleteAttempts int     `json:"failedDeleteAttempts"`
	Deleted              *string `json:"deleted"`
}

func changeToExeDir() error {
	exePath, _ := os.Executable()
	exeDir, _ := filepath.Split(exePath)
	if exeDir != "" {
		return os.Chdir(exeDir)
	}
	return nil
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func logActivity(programName, message string) error {
	f, err := os.OpenFile(ActivityLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(nowUTC() + " - " + programName + " - " + message + "\n")
	return err
}

func loadPendingItems() ([]PendingItem, error) {
	file, err := os.Open(PendingItemsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []PendingItem{}, nil
		}
		return nil, err
	}
	defer file.Close()

	var items []PendingItem
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var item PendingItem
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, scanner.Err()
}

func savePendingItems(items []PendingItem) error {
	file, err := os.Create(PendingItemsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		if _, err := file.WriteString(string(data) + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func itemExists(items []PendingItem, path string) int {
	for i, item := range items {
		if item.Path == path {
			return i
		}
	}
	return -1
}

func absPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	abs, _ := filepath.Abs(path)
	return abs
}

func showNotification(title, body string) error {
	notification := toast.Notification{
		Title:   title,
		Message: body,
	}
	return notification.Push()
}

func main() {
	changeToExeDir()

	if len(os.Args) < 2 {
		logActivity("DUHKHA", "no path provided")
		return
	}

	path := absPath(os.Args[1])

	if err := showNotification("Impermanence", "Will delete in 30 days:\n"+path); err != nil {
		logActivity("DUHKHA", "notification error: "+err.Error())
	}

	items, err := loadPendingItems()
	if err != nil {
		logActivity("DUHKHA", "error loading pending items: "+err.Error())
		return
	}

	idx := itemExists(items, path)
	if idx >= 0 && items[idx].Deleted != nil {
		deleteAt := time.Now().UTC().AddDate(0, 0, DeleteDelayDays).Format(time.RFC3339)
		items[idx].DeleteAt = deleteAt
		if err := savePendingItems(items); err != nil {
			logActivity("DUHKHA", "error updating existing item: "+err.Error())
			return
		}
		logActivity("DUHKHA", "updated deleteAt for existing item: "+path)
		return
	}

	item := PendingItem{
		Added:                nowUTC(),
		Path:                 path,
		Computer:             ComputerName,
		DeleteAt:             time.Now().UTC().AddDate(0, 0, DeleteDelayDays).Format(time.RFC3339),
		FailedDeleteAttempts: 0,
		Deleted:              nil,
	}

	items = append(items, item)

	if err := savePendingItems(items); err != nil {
		logActivity("DUHKHA", "error saving pending items: "+err.Error())
		return
	}

	logActivity("DUHKHA", "added: "+path)
}