package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Log writes error to log file and displays short error to display or writes
// full error directly to screen if error happend when writing data to file
//
// It accepts error and short explanation message which will be displayed on
// screen as well as written to log file
func Log(err error, text string) {
	fmt.Println("\x1B[31mError occured when trying to: " + text + "\x1B[0m")
	writeLog(text + "\n" + err.Error())
}

// Log writes error to log file and displays short error to display or writes
// full error directly to screen if error happend when writing data to file.
// After that it exits programm with error code
//
// It accepts error and short explanation message which will be displayed on
// screen as well as written to log file
func LogCritical(err error, text string) {
	fmt.Println("\x1B[31mError occured when trying to: " + text + "\x1B[0m")
	writeLog(text + " : " + err.Error())
	os.Exit(1)
}

// writeLog creates log direcotry if it is not exists and writes data provided
// by Log or LogCritical to file
//
// It accepts string which will be written to log file
func writeLog(text string) {
	if _, err := os.Stat(Config.LogsDir); os.IsNotExist(err) {
		err = os.Mkdir(Config.LogsDir, 0777)
		if err != nil {
			fmt.Printf("\x1B[31mError occured when trying to create log directory: %v\nError text: %v\n\x1B[0m", err.Error(), text)
			return
		}
	}

	fName := fmt.Sprintf("%v/%v.log", Config.LogsDir, time.Now().Format("2006-01-02-15:04:05"))
	err := ioutil.WriteFile(fName, []byte(text), 0777)
	if err != nil {
		fmt.Printf("\x1B[31mError occured when trying to write log file: %v\nError text: %v\n\x1B[0m", err.Error(), text)
	}
}
