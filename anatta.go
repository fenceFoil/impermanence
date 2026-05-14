package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	PendingItemsFile = "data/pendingItems.jsonl"
	ActivityLogFile  = "data/activityLog.txt"
	ComputerName     = "VALJEAN"
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

func deletePath(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}

	if info.IsDir() {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}

	if err != nil {
		_, statErr := os.Stat(path)
		if statErr == nil {
			return false
		}
		return true
	}

	_, err = os.Stat(path)
	return os.IsNotExist(err)
}

func main() {
	changeToExeDir()
	logActivity("ANATTA", "started")

	items, err := loadPendingItems()
	if err != nil {
		logActivity("ANATTA", "error loading pending items: "+err.Error())
		return
	}

	for i := range items {
		item := &items[i]
		if item.Computer != ComputerName {
			continue
		}
		if item.Deleted != nil {
			continue
		}

		deleteAt, err := time.Parse(time.RFC3339, item.DeleteAt)
		if err != nil {
			logActivity("ANATTA", "invalid deleteAt for item: "+item.Path)
			continue
		}

		if deleteAt.After(time.Now().UTC()) {
			continue
		}

		success := deletePath(item.Path)
		if success {
			deletedAt := nowUTC()
			item.Deleted = &deletedAt
			logActivity("ANATTA", "deleted: "+item.Path)
		} else {
			item.FailedDeleteAttempts++
			logActivity("ANATTA", "failed to delete: "+item.Path)
		}

		if err := savePendingItems(items); err != nil {
			logActivity("ANATTA", "error saving pending items: "+err.Error())
		}
	}
}