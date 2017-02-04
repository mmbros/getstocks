package main

import (
	"os"

	"github.com/mmbros/getstocks/cli"
	log "github.com/sirupsen/logrus"
)

func initLog(filename string) *os.File {
	path := filename
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0640)
	if err != nil {
		log.WithFields(log.Fields{
			"path":  path,
			"error": err.Error(),
		}).Error("Can't open log file")
		return nil
	}
	log.SetOutput(f)
	log.SetFormatter(&log.JSONFormatter{})
	return f
}

func main() {
	logfile := initLog("getstocks.log")
	log.Infof("PROGRAM START")
	res := cli.Run()
	log.Infof("PROGRAM FINISH")
	if logfile != nil {
		logfile.Close()
	}
	os.Exit(res)
}
