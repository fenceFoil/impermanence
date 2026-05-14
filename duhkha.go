package main

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
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

const (
	MB_OK              = 0x00000000
	MB_ICONINFORMATION = 0x00000040
	MB_SETFOREGROUND   = 0x00010000
	MB_TOPMOST         = 0x00040000
	IDTIMEOUT          = 32000
)

var user32 = syscall.MustLoadDLL("user32.dll")
var messageBoxW = user32.MustFindProc("MessageBoxW")

func messageBox(hwnd uintptr, text, caption *uint16, flags uint32) int {
	ret, _, _ := messageBoxW.Call(hwnd, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(caption)), uintptr(flags))
	return int(ret)
}

func showConfirmDialog(path string) int {
	text, _ := syscall.UTF16PtrFromString("This item will be deleted in 30 days:\n\n" + path)
	caption, _ := syscall.UTF16PtrFromString("Impermanence")

	done := make(chan int)
	go func() {
		done <- messageBox(0, text, caption, MB_OK|MB_ICONINFORMATION|MB_SETFOREGROUND|MB_TOPMOST)
	}()

	select {
	case result := <-done:
		return result
	case <-time.After(3 * time.Second):
		return IDTIMEOUT
	}
}

func main() {
	changeToExeDir()

	if len(os.Args) < 2 {
		logActivity("DUHKHA", "no path provided")
		return
	}

	path := absPath(os.Args[1])

	result := showConfirmDialog(path)
	if result == 0 || result == IDTIMEOUT {
		logActivity("DUHKHA", "dialog timeout/cancel for: "+path)
		return
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
