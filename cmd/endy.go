package main

import (
	"flag"
	"time"

	"github.com/gonnafaraway/endy"
)

const (
	PathFlag    = "path"
	TimeoutFlag = "timeout"

	DefaultTimeout = 10 * time.Second
)

func main() {
	pflag := flag.String(PathFlag, "config.yaml", "path to config file with end-to-end cases")
	tflag := flag.Duration(TimeoutFlag, DefaultTimeout, "limit of all tests duration")

	t := endy.New()

	cfg := endy.Config{
		Path:    *pflag,
		Timeout: *tflag,
	}

	t.Config = &cfg

	err := t.Run()
	if err != nil {
		t.Logger.Fatal(err.Error())
	}
}
