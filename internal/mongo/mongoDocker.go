package mongo

import (
	"os/exec"

	"github.com/automotiveMastermind/condo-cli/internal/docker"
	log "github.com/sirupsen/logrus"
)

var MONGO_INSTANCE_NAME string = "mongo-container"

// Run mongo on docker and connect it to the clusters network
func Run() {

	if docker.IsImageRunning(MONGO_INSTANCE_NAME) {
		log.Info(MONGO_INSTANCE_NAME + " is already running, skipping " + MONGO_INSTANCE_NAME + " creation.")
		return
	}

	log.Info("Starting " + MONGO_INSTANCE_NAME)
	cmd := exec.Command(
		"docker",
		"run",
		"-d",
		"-p27017:27017",
		"--pull=missing",
		"--name="+MONGO_INSTANCE_NAME,
		"--restart=always",
		"mongo",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Infof("%s", out)
		log.Infof("failed to start "+MONGO_INSTANCE_NAME+": %v", err)

	}

	// connect to kind network
	cmd = exec.Command(
		"docker",
		"network",
		"connect",
		"kind",
		MONGO_INSTANCE_NAME,
	)

	out, err = cmd.CombinedOutput()
	if err != nil {
		log.Infof("%s", out)
		log.Fatalf("failed to add "+MONGO_INSTANCE_NAME+" to the kind network: %v", err)
	}
	log.Info("attach " + MONGO_INSTANCE_NAME + "to kind network")

}

// Stop mongo running on docker
func Stop() {
	log.Info("Removing container " + MONGO_INSTANCE_NAME + " from docker")

	//stop the mongo container
	dockerStopCmd := exec.Command(
		"docker",
		"stop",
		MONGO_INSTANCE_NAME,
	)

	mongoStopErr := dockerStopCmd.Run()
	if mongoStopErr != nil {

		log.Infof("Docker container \""+MONGO_INSTANCE_NAME+"\" failed to stop:  %v", mongoStopErr)
	}

	//remove the mongo container
	dockerRemoveCmd := exec.Command(
		"docker",
		"rm",
		MONGO_INSTANCE_NAME,
	)

	dockerRemoveErr := dockerRemoveCmd.Run()
	if dockerRemoveErr != nil {

		log.Infof("Docker container \""+MONGO_INSTANCE_NAME+"\" failed to be removed:  %v", dockerRemoveErr)
	}

	log.Info(MONGO_INSTANCE_NAME + " removed from docker")

}
