# automate

A utility for easily scheduling tasks in Windows (actually, it is cross-platform but cron is a better alternative on other platforms) written in Go. Gorma runs entirely in the background, with the exception of an icon in the tray (to show that it is active). 

**Currently, only the following functionality is available:**
On a monthly, weekly, daily, hourly, and per minute basis, the application attempts to execute all files in the folders with corresponding names (month/week/day/hour/minute).

Commands run on startup, then on the hour for hourly tasks and at midnight for weekly and daily dasks.

In the future, I may implement reading a crontab-file.