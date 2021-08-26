package services

import (
	"os/exec"

	log "github.com/sirupsen/logrus"
)

var MONGO_INSTANCE_NAME string = "mongo-container"

func checkMongoRunning() bool {

	cmd := exec.Command(
		"docker",
		"container",
		"inspect",
		"-f",
		"'{{.State.Running}}'",
		MONGO_INSTANCE_NAME,
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

func InstallMongo() {

	if checkDockerRegistryRunning() {
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

func RemoveMongoDockerContainer() {
	log.Info("Removing container " + MONGO_INSTANCE_NAME + " from docker")

	//stop the mongo container
	dockerStopCmd := exec.Command(
		"docker",
		"stop",
		MONGO_INSTANCE_NAME,
	)
	var mongoStopErr error
	mongoStopErr = dockerStopCmd.Run()
	if mongoStopErr != nil {

		log.Infof("Docker container \""+MONGO_INSTANCE_NAME+"\" failed to stop:  %v", mongoStopErr)
	}

	//remove the mongo container
	dockerRemoveCmd := exec.Command(
		"docker",
		"rm",
		MONGO_INSTANCE_NAME,
	)
	var dockerRemoveErr error
	dockerRemoveErr = dockerRemoveCmd.Run()
	if dockerRemoveErr != nil {

		log.Infof("Docker container \""+MONGO_INSTANCE_NAME+"\" failed to be removed:  %v", dockerRemoveErr)
	}

	log.Info(MONGO_INSTANCE_NAME + " removed from docker")

}
