package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	filenames := flag.String("file", "", "log file names delimited by comas")
	format := flag.String("format", "", "log file format")
	config := flag.String("config", "config.json", "configuration file for database connection")

	flag.Parse()

	if len(*format) == 0 {
		log.Fatal(errors.New("input log format"))
	}
	if len(*filenames) == 0 {
		log.Fatal(errors.New("input file names"))
	}
	filenamesList := strings.Split(*filenames, ",")
	filenamesList, err := Map(filenamesList, filepath.Abs)
	if err != nil {
		log.Fatal(err)
	}

	err = dbConnect(*config)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Files to process: ", filenamesList)
	fmt.Println("Log format: ", *format)
	fmt.Println("\nhint: input exit to quit")

	exit := make(chan string)
	go listenForExit(exit)

	go createRoutinesAndReceiveLogs(&filenamesList, *format)

	for {
		select {
		case msg := <-exit:
			if msg == "exit" {
				return
			}
		}
	}
}

func listenForExit(exitChan chan string) {
	reader := bufio.NewReader(os.Stdin)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if strings.TrimSpace(msg) == "exit" {
			exitChan <- "exit"
		}
	}
}

func createRoutinesAndReceiveLogs(fileList *[]string, fileFormat string)  {
	objChan := make(chan *Log)
	for _, fileName := range *fileList {
		go parseLogFile(fileName, fileFormat, objChan)
	}

	for {
		select {
		case logObj := <-objChan:
			dbSaveLog(logObj)
		}
	}
}

//Map applies given function to the passed slice;
// returns the result and error
func Map(vs []string, f func(string) (string, error)) ([]string, error) {
	var err error
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i], err = f(v)
		if err != nil {
			return vsm, err
		}
	}
	return vsm, err
}
