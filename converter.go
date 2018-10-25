package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func parseLogFile(fileName, fileFormat string, logObjChan chan *Log) {
	file, oldFileSize, err := openFileGetSize(fileName)
	if err != nil {
		fmt.Print(err)
		return
	}

	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {
		if len(fileScanner.Text()) == 0 {
			continue
		}
		logObj, err := readFormattedLog(fileScanner.Text(), file.Name(), fileFormat)
		if err != nil {
			file.Close()
			fmt.Print(err)
			return
		}
		logObjChan <- logObj
	}
	file.Close()

	err = watchLogFile(fileName, fileFormat, oldFileSize, logObjChan)
	if err != nil {
		fmt.Print(err)
	}
	return
}

func openFileGetSize(fileName string) (*os.File, int64, error) {
	file, err := os.Open(fileName)
	if err != nil {
		file.Close()
		return nil, 0, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, err
	}

	return file, fileInfo.Size(), nil
}

func readFormattedLog(fileRow, fileName, fileFormat string) (*Log, error) {
	var newLog Log
	var err error

	rowData := strings.Split(fileRow, " | ")

	switch fileFormat {

	case "first_format":
		layout := "Jan 2, 2006 at 3:04:05pm (UTC)"

		newLog.LogTime, err = time.Parse(layout, rowData[0])
		if err != nil {
			return nil, err
		}

	case "second_format":
		newLog.LogTime, err = time.Parse(time.RFC3339, rowData[0])
		if err != nil {
			return nil, err
		}

	default:
		return nil, errors.New("unknown log format")
	}
	newLog.LogMsg = rowData[1]
	newLog.LogFormat = fileFormat
	newLog.FileName = fileName

	return &newLog, nil
}

func watchLogFile(fileName, fileFormat string, oldFileSize int64, logObjChan chan *Log) error {
	ticker := time.Tick(2 * time.Second)

	for {
		select {
		case <-ticker:

			newFile, newFileSize, err := openFileGetSize(fileName)
			if err != nil {
				return err
			}

			if newFileSize == oldFileSize {
				newFile.Close()
				continue
			}

			newFileReader, err := offsetFileGetReader(oldFileSize, newFile)
			if err != nil {
				newFile.Close()
				return err
			}

			err = readAndSendFileNewData(newFileReader, newFile, fileFormat, logObjChan)
			if err != nil {
				newFile.Close()
				return err
			}

			newFile.Close()
			oldFileSize = newFileSize
		}
	}

	return nil
}

func offsetFileGetReader(offset int64, file *os.File) (*bufio.Reader, error) {
	_, err := file.Seek(offset, 0)
	if err != nil {
		return nil, err
	}
	fileReader := bufio.NewReader(file)
	return fileReader, nil
}

func readAndSendFileNewData(newFileReader *bufio.Reader,
	newFile *os.File, fileFormat string, logObjChan chan *Log) error {

	for {
		line, err := newFileReader.ReadBytes('\n')

		if err == nil || err == io.EOF {
			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			var logObj *Log
			if len(line) != 0 {
				logObj, err = readFormattedLog(string(line), newFile.Name(), fileFormat)
				if err != nil {
					newFile.Close()
					return err
				}
			}

			logObjChan <- logObj
		}
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
	}
	return nil
}
