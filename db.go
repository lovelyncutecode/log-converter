package main

import (
	"gopkg.in/mgo.v2"
	"path/filepath"
)

var logCollection *mgo.Collection

func dbConnect(configFile string) error {
	absConfigFile, err := filepath.Abs(configFile)
	if err != nil {
		return err
	}
	config, err := getConfigVars(absConfigFile)
	if err != nil {
		return err
	}

	session, err := mgo.Dial("mongodb://" + config.DBHost)
	if err != nil {
		return err
	}
	logCollection = session.DB(config.DBName).C(config.DBCollection)

	return nil
}

func dbSaveLog(log *Log) error {
	err := logCollection.Insert(log)
	if err != nil {
		return err
	}
	return nil
}
