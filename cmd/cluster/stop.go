package cluster

import (
	"github.com/automotiveMastermind/condo-cli/internal/docker"
	"github.com/automotiveMastermind/condo-cli/internal/git"
	"github.com/automotiveMastermind/condo-cli/internal/mongo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	stopLong string = `Stops the kubernetes cluster on your local docker instance but maintains the configuration
    and changes you have made in your clusters folder`
)

func NewCmdStop() *cobra.Command {
	// clusterCmd represents the cluster command specific to the stop command
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stops the kubernetes cluster on your local docker instance",
		Long:  stopLong,
		Run: func(cmd *cobra.Command, args []string) {
			stopCluster()
		},
	}

	return cmd
}

func stopCluster() {
	docker.IsRunning()
	//clusterExistCheck()
	git.Stop()
	docker.Stop()
	mongo.Stop()
	//removeClusterNodes()

	log.Info("Cluster stopped")
}
