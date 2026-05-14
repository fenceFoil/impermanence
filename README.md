# Impermanence

A file deletion scheduling system for Windows.

## Overview

Two programs manage pending file deletions:

- **anatta.exe** - Deletes files on schedule. Runs daily via Task Scheduler and on startup.
- **duhkha.exe** - Adds files to the deletion queue. Accessible via Windows Explorer right-click menu.

## Architecture

```
                    pendingItems.jsonl
                         ^
        +-----------------|
        |                 |
   [Right-click]    [Scheduled]
        |                 |
        v                 v 
    duhkha.exe -----> anatta.exe --------> (delete files)
        adds             deletes           permanently
```

## Installation

1. Run `build.bat` to build the executables
2. Run `AddDuhkhaToRightClickMenus.reg` to add "Delete Next Month" to context menus
3. Run `makeRunDailyAndAtStartup.bat` to schedule anatta.exe to run daily and on startup