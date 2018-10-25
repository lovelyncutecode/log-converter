package main

import "time"

type Log struct {
	LogTime   time.Time `bson:"log_time"`
	LogMsg    string    `bson:"log_msg"`
	FileName  string    `bson:"file_name"`
	LogFormat string    `bson:"log_format"`
}
