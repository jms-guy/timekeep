## Commands for CLI Use

- `add`
    - Add a program to begin tracking. Add name of program's executable file name. May specify any number of programs to track in a single command, seperated by spaces in between
    - `timekeep add notepad.exe`, `timekeep add notepad.exe code.exe chrome.exe`

- `rm`
    - Remove a program from tracking list. May specify any number of programs to remove in a single command, seperated by spaces in between. Takes `--all` flag to clear program list completely
    - `timekeep rm notepad.exe`, `timekeep rm --all`
    
- `ls`
    - Lists programs being tracked by service
    - `timekeep ls`

- `info`
    - Shows basic info for currently tracked programs. Accepts program name as argument to show in-depth stats for that program, else shows basic stats for all programs
    - `timekeep info`, `timekeep info notepad.exe`

- `active`
    - Display list of current active sessions being tracked by service
    - `timekeep active`

- `history`
    - Shows session history, may take program name as argument to filter sessions shown
    - `timekeep history`, `timekeep history notepad.exe`
    - Flags available for further filtering:
        - ex. `timekeep history --date 2025-09-30 --limit 10`
        - `date` (2006-01-02) - Show sessions open on given date
        - `start` (2006-01-02) - Show sessions open on or after given date
        - `end` (2006-01-02) - If flag is given alongside `start`, will filter sessions open up-to given date
        - `limit` (25) - Will specify number of sessions to show at one time. Default 25 
    

- `refresh`
    - Sends a manual refresh command to the service
    - `timekeep refresh`

- `reset`
    - Reset tracking stats for given programs. Accepts multiple arguments seperated by space. Takes `--all` flag to reset all stats
    - `timekeep reset notepad.exe`, `timekeep reset --all`

- `status`
    - Gets current state of Timekeep service
    - `timekeep ping`

- `version`
    - Returns version of Timekeep user is running