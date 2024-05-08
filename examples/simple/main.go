package main

import (
	"fmt"
	"github.com/gonnafaraway/endy"
	"os"
	"time"
)

func main() {
	e := endy.New()

	e.SetTimeout(10 * time.Second)

	e.SetConfigPath("config.yaml")

	if err := e.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
