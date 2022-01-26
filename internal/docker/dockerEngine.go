package docker

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

//check if docker engine is running
func IsRunning() {

	log.Info("Checking if docker daemon is running...")
	cmd := exec.Command("docker", "ps")

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Docker daemon was not found:  %v", err)
	}

	log.Info("Docker daemon running")
}

func IsImageRunning(name string) bool {
	cmd := exec.Command(
		"docker",
		"container",
		"inspect",
		"-f",
		"'{{.State.Running}}'",
		name,
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("%s", err)
		return false
	}

	if string(out) == "true" {
		return true
	} else {
		return false
	}
}
