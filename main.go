package main

import (
	"log"

	"github.com/distek/tmux-tools/cmd"
)

func main() {
	log.SetFlags(log.Lshortfile)
	cmd.Execute()
}
