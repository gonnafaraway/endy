package main

import (
	"flag"
	"time"

	"github.com/gonnafaraway/endy"
)

const (
	PathFlag    = "path"
	TimeoutFlag = "timeout"
	BenchFlag   = "bench"

	DefaultTimeout = 10 * time.Second
)

func main() {
	pflag := flag.String(PathFlag, "config.yaml", "path to config file with end-to-end cases")
	tflag := flag.Duration(TimeoutFlag, DefaultTimeout, "limit of all tests duration")
	bflag := flag.Bool(BenchFlag, false, "execute tests in benchmark mode")

	flag.Parse()

	t := endy.New()

	cfg := endy.Config{
		Path:      *pflag,
		Timeout:   *tflag,
		BenchMode: *bflag,
	}

	t.Config = &cfg

	err := t.Run()
	if err != nil {
		t.Logger.Fatal(err.Error())
	}
}
