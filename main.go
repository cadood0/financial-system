package main

import (
	"financial-system/cmd"
	"financial-system/config"
)

func main() {
	cfg := config.Load()
	cmd.Start(cfg)
}
