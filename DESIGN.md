# Impermanence Design

## Overall

### Notes For All Programs

All programs will be written in Go and compiled for windows and linux.

A batch file, build.bat, should be created that does that windows & linux build.

All programs will not launch with the current working directory set to this project folder. After being launched, the programs should set their jown current working directory to the location of the executable.

### The Pending Items Database

A jsonl file, stored locally (in this project folder), that is a list of items to delete. Name: data/pendingItems.jsonl

For each item's pending item json object, the following keys are saved:

- added: iso utc timestamp string
- path: File/folder path string
- computer: always "VALJEAN" for now, string
- deleteAt: iso utc timestamp string
- failedDeleteAttempts: a number that starts at 0
- deleted: starts as null, when deleted populate with iso utc timestamp string

### The Activity Log

A text log file stored locally in this project folder named data/activityLog.txt

Every time a program writes to it, prepend an ISO timestamp in UTC and a hyphen and the name of the script (i.e. DUHKHA and ANATTA)

## File Deleting Script - anatta.exe

Go program designed to be run once a day by the windows task scheduler, and on startup.

On launch, log a note into the activity log saying ANATTA started

When run, it reviews each of the pending items in the pending item database. Only process ones with computer == "VALJEAN". If the deleteAt time is now passed, attempt to delete the file or folder at path recursively and permanently. If this fails, or a double-check after trying to delete the item shows it is still there, increment the failedDeleteAttempts and update the pending items database file. If the deletion succeeds, set deletedAt and deleted and save.

Resave the pendingItems.jsonl file after each item is handled: don't wait until end to rewrite once.

After deleting or failing to delete an item, log a note into the activity log with what happened and the path of the item being operated on.

## Item Adder - duhkha.exe

Go program that adds a new path to the end of the pendingItems.jsonl, with the added and path and computer and deleteAt fields filled out. deleteAt is set to a month in the future. deleted and failedDeleteAttempts are null and 0, respectively.

If the path already exists in the pendingItems.jsonl and the value of that duplicate's .deleted is non-null, only update the deleteAt string of that existing item, don't add a new one.

Once the item is added, put a note into the activity log that it was added, and its path.

Duhkha is designed to be launched from the windows right-click menu in explorer. It accepts one argument: the path to be deleted.

An install script, "AddDuhkhaToRightClickMenus.reg", adds duhkha.exe to the right click menus in explorer for both files and directories/folders. The action will be called "ImpermanenceDuhkha" and the text of the action in the menu will be "Delete Next Month".