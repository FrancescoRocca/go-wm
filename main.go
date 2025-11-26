package main

import (
	"fmt"
	"gowm/wm"
	"log"
	"os"
)

func main() {
	log.Println("Init Window Manager")
	windowmanager, err := wm.InitWM()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	log.Println("Run WM")
	windowmanager.Run()

	log.Println("Close WM")
	windowmanager.Close()
}
