package main

import (
	"encoding/csv"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/getlantern/systray"
)

var pCrontab *regexp.Regexp = regexp.MustCompile(`^(\*(?:\/[0-9]+)?|[0-9]+(?:-[0-9]+|(?:,[0-9]+)+)?)\s+(\*(?:\/[0-9]+)?|[0-9]+(?:-[0-9]+|(?:,[0-9]+)+)?)\s+(\*(?:\/[0-9]+)?|[0-9]+(?:-[0-9]+|(?:,[0-9]+)+)?)\s+(\*(?:\/[0-9]+)?|[0-9]+(?:-[0-9]+|(?:,[0-9]+)+)?)\s+(\*(?:\/[0-9]+)?|[0-9]+(?:-[0-9]+|(?:,[0-9]+)+)?)\s+(.+)$`)
var pFraction *regexp.Regexp = regexp.MustCompile(`^\*\/([0-9]+)$`)
var pRange *regexp.Regexp = regexp.MustCompile(`^([0-9]+)-([0-9]+)$`)
var pList *regexp.Regexp = regexp.MustCompile(`^[0-9]+(?:,[0-9]+)*$`)

func main() {
	// Should be called at the very beginning of main().
	systray.Run(onReady, onExit)
}

func listDir(path string) ([]fs.FileInfo, error) {
	d, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer d.Close()

	files, err := d.Readdir(0)
	if err != nil {
		return nil, err
	}

	return files, nil
}

func runTask(task string, args ...string) {
	cmd := exec.Command(task, args...)
	log.Printf("executing [%s %v]", task, args)
	err := cmd.Run()
	if err != nil {
		log.Printf("ERROR - [%s %v] finished with error: %v", task, args, err)
	} else {
		log.Printf("[%s %v] finished without errors.", task, args)
	}
}

func startTasker(taskdir string, interval time.Duration) {
	_ = os.Mkdir(taskdir, os.ModePerm) //Create the dir, if it doesnt exist

	for {
		wd, err := os.Getwd()
		if err != nil {
			log.Printf("ERROR getting working directory: %v", err)
			continue
		}

		files, err := listDir(taskdir)

		if err != nil {
			log.Printf("error listing files in '%v': %v", taskdir, err)
		} else {
			for _, f := range files {
				file := path.Join(wd, taskdir, f.Name())
				go runTask(file)
			}
		}

		nextTime := time.Now().Truncate(interval)
		nextTime = nextTime.Add(interval)

		time.Sleep(time.Until(nextTime))
	}
}

type tabTask struct {
	Minute    bool
	Hour      bool
	Day       bool
	Month     bool
	DOW       bool
	Command   string
	Arguments []string
}

func parseTabTaskTag(tag string, currentValue int) bool {
	if tag == "*" {
		return true
	}

	if m := pFraction.FindStringSubmatch(tag); len(m) == 2 {
		f, _ := strconv.Atoi(m[1])

		if f != 0 && currentValue%f == 0 {
			return true
		}

		return false
	}

	if m := pRange.FindStringSubmatch(tag); len(m) == 3 {
		min, _ := strconv.Atoi(m[1])
		max, _ := strconv.Atoi(m[2])

		if currentValue >= min && currentValue <= max {
			return true
		}

		return false
	}

	if pList.MatchString(tag) {
		alts := strings.Split(tag, ",")

		for _, a := range alts {
			v, _ := strconv.Atoi(a)

			if v == currentValue {
				return true
			}
		}
	}

	return false
}

func parseTabTask(forTime time.Time, qMinute, qHour, qDay, qMonth, qDOW, cmd string) *tabTask {
	hour, minute, _ := forTime.Clock()
	_, tMonth, day := forTime.Date()
	month := int(tMonth)
	dow := int(forTime.Weekday())

	task := tabTask{
		Minute: parseTabTaskTag(qMinute, minute),
		Hour:   parseTabTaskTag(qHour, hour),
		Day:    parseTabTaskTag(qDay, day),
		Month:  parseTabTaskTag(qMonth, month),
		DOW:    parseTabTaskTag(qDOW, dow),
	}

	r := csv.NewReader(strings.NewReader(cmd))
	r.Comma = ' '
	fields, err := r.Read()
	if err != nil {
		log.Printf("error while reading command [%s] in crontab.txt: '%v'", cmd, err)
		return nil
	}

	task.Command = fields[0]
	if len(fields) > 1 {
		task.Arguments = fields[1:]
	} else {
		task.Arguments = make([]string, 0)
	}

	return &task
}

func tabTasks() {
	for {
		currentTime := time.Now().Truncate(time.Minute)

		f, err := os.OpenFile("crontab.txt", os.O_CREATE, 0666)
		if err != nil {
			log.Printf("error opening crontab.txt: '%v'", err)
			return
		}

		content, err := io.ReadAll(f)
		if err != nil {
			log.Printf("error reading crontab.txt: '%v'", err)
		}
		f.Close()

		lines := strings.Split(string(content), "\n")
		for _, l := range lines {
			m := pCrontab.FindStringSubmatch(l)

			if len(m) < 7 {
				continue
			}

			task := parseTabTask(currentTime, m[1], m[2], m[3], m[4], m[5], m[6])
			if task != nil && task.Minute && task.Hour && task.Day && task.Month && task.DOW {
				go runTask(task.Command, task.Arguments...)
			}
		}

		nextTime := currentTime.Add(time.Minute)
		time.Sleep(time.Until(nextTime))
	}
}

func setUpTasks() {
	go startTasker("month", 7*24*30*time.Hour)
	go startTasker("week", 7*24*time.Hour)
	go startTasker("day", 24*time.Hour)
	go startTasker("hour", time.Hour)
	go startTasker("minute", time.Minute)

	go tabTasks()

	log.Printf("start-up complete")
}

func onReady() {
	systray.SetIcon(icon)
	systray.SetTooltip("Running scheduled tasks")
	mQuit := systray.AddMenuItem("Quit", "Close the application")

	f, err := os.OpenFile("logfile.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	} else {
		log.SetOutput(f)
	}
	defer f.Close()

	setUpTasks()

	<-mQuit.ClickedCh
	f.Close()
	systray.Quit()
}

func onExit() {
	os.Exit(0)
}
