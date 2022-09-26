# Automate

Automate is a basic and light-weight utility for easily scheduling tasks in Windows (actually, it is cross-platform but I see no reason why anyone wouldnt use a more feature rich and well documented cron daemon on e.g. Linux) written in Go. Automate runs entirely in the background, with the exception of an icon in the tray (to show that it is active). I am aware of the windows task scheduler, but it has never done anything for me other than contribute to my receeding hairline.

## Folder based automation

On a monthly, weekly, daily, hourly, and per minute basis, the application attempts to execute all files in the folders with corresponding names (month/week/day/hour/minute), empty folders are created in the working directory on launch if they do not exist.

Commands run on startup, then on the hour for hourly tasks and at midnight for weekly and daily dasks.

## Crontab.txt

Additionally, you may write jobs in a file name crontab.txt inside the working directory of automate (created on launch if it does not exist). Each line in crontab.txt should begin with 5 time tags separated with whitespace followed by the command to execute, like below:

```
.---------------- minute
| .-------------- hour
| | .------------ day of month
| | | .---------- month
| | | | .-------- day of week
| | | | |
* * * * * your-command arg1 arg2 arg3...
```

Day of week is given as a number between 1 and 7, where 1 is monday and 7 is sunday. The crontab.txt file is read once per minute. Examples of valid values in time tags are:

| Value | Type            | Description                                                                                                 |
|---    |---              |---                                                                                                          |
| *     | Any             | Execute for any value in this category                                                                      |
| 5     | Single value    | Execute when the value for that time-tag matches the value                                                  |
| 1,5,9 | Multiple values | Execute when the value for that time-tag matches any value in the comma separated list                      |
| 1-9   | Range           | Execute when the value for that time-tag falls inside the specified range                                   |
| */5   | Fraction        | Execute when the value for that time-tag is divisible according to the given fraction (e.g every 5 minutes) |

### Examples

#### Run a job at 07:30

```
30 7 * * * command
```

#### Run a job at midning at the first day of every month

```
* 0 1 * * command
```

#### Run a job every 5 minutes

```
*/5 * * * * command
```
