## Commands for CLI Use

- `active`
    - Display list of current active sessions being tracked by service
    - `timekeep active`

- `add`
    - Add a program to begin tracking. Add name of program's executable file name. May specify any number of programs to track in a single command, seperated by spaces in between
    - `timekeep add notepad.exe`, `timekeep add notepad.exe code.exe chrome.exe`
    - Flags available:
        - `category` - Set category for program, required for WakaTime tracking (`timekeep add notepad.exe --category notes`)
        - `project` - Set project for WakaTime data sorting (`timekeep add notepad.exe --category notes --project timekeep`)

- `history`
    - Shows session history, may take program name as argument to filter sessions shown
    - `timekeep history`, `timekeep history notepad.exe`
    - Flags available for further filtering:
        - ex. `timekeep history --date 2025-09-30 --limit 10`
        - `date` (2006-01-02) - Show sessions open on given date
        - `start` (2006-01-02) - Show sessions open on or after given date
        - `end` (2006-01-02) - If flag is given alongside `start`, will filter sessions open up-to given date
        - `limit` (25) - Will specify number of sessions to show at one time. Default 25 
    
- `info`
    - Shows basic info for currently tracked programs. Accepts program name as argument to show in-depth stats for that program, else shows basic stats for all programs
    - `timekeep info`, `timekeep info notepad.exe`
    
- `ls`
    - Lists programs being tracked by service
    - `timekeep ls`

- `refresh`
    - Sends a manual refresh command to the service
    - `timekeep refresh`

- `reset`
    - Reset tracking stats for given programs. Accepts multiple arguments seperated by space. Takes `--all` flag to reset all stats
    - `timekeep reset notepad.exe`, `timekeep reset --all`

- `rm`
    - Remove a program from tracking list. May specify any number of programs to remove in a single command, seperated by spaces in between. Takes `--all` flag to clear program list completely
    - `timekeep rm notepad.exe`, `timekeep rm --all`

- `status`
    - Gets current state of Timekeep service
    - `timekeep ping`

- `update`
    - Update a given program's category/project fields
    - Flags for each field:
        - `--category`, `--project`
    - `timekeep update notepad.exe --category coding --project testing`

- `version`
    - Returns version of Timekeep user is running

- `wakatime [status|enable|disable|set-path|set-project]`
    - Enable WakaTime integration with `timekeep wakatime enable`
        - Flags:
            - `--api-key "KEY"` - Set WakaTime API key
            - `--set-path "PATH"` - Set wakatime-cli path(absolute)
    - Disable integration with `timekeep wakatime disable`
    - Check WakaTime enabled/disabled status with `timekeep wakatime status`
    - Set wakatime-cli path with command `timekeep wakatime set-path "PATH"`
    - Set global_project config variable with `timekeep wakatime set-project "YOUR_PROJECT"`