package wm

import (
	"log"
	"os/exec"
)

func spawnLauncher() error {
	commands := [][]string{
		{"dmenu_run"},
		{"rofi", "-show", "run"},
		{"xterm"},
	}

	var lastErr error

	for _, cmd := range commands {
		if len(cmd) == 0 {
			continue
		}

		if _, err := exec.LookPath(cmd[0]); err != nil {
			lastErr = err
			log.Printf("spawnLauncher: %s not found: %v\n", cmd[0], err)
			continue
		}

		if err := exec.Command(cmd[0], cmd[1:]...).Start(); err == nil {
			return nil
		} else {
			lastErr = err
			log.Printf("spawnLauncher: error launching %v: %v\n", cmd, err)
		}
	}

	return lastErr
}

func (wm *WindowManager) runLauncher() {
	if err := spawnLauncher(); err != nil {
		log.Println("Spawn error:", err)
	}
}
