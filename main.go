package main

import (
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path"
	"time"

	"github.com/getlantern/systray"
)

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

func runTask(task string) {
	cmd := exec.Command(task)
	log.Printf("executing [%s]", task)
	err := cmd.Run()
	if err != nil {
		log.Printf("ERROR - [%s] finished with error: %v", task, err)
	} else {
		log.Printf("[%s] finished without errors.", task)
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

func setUpTasks() {
	go startTasker("month", 7*24*30*time.Hour)
	go startTasker("week", 7*24*time.Hour)
	go startTasker("day", 24*time.Hour)
	go startTasker("hour", time.Hour)
	go startTasker("minute", time.Minute)

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

	setUpTasks()

	<-mQuit.ClickedCh
	f.Close()
	systray.Quit()
}

func onExit() {
	os.Exit(0)
}
