package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"path/filepath"

	"github.com/automotiveMastermind/condo-cli/cmd/cluster"
	"github.com/automotiveMastermind/condo-cli/internal/console"
	"github.com/automotiveMastermind/condo-cli/internal/docker"
	"github.com/automotiveMastermind/condo-cli/internal/flux"
	"github.com/automotiveMastermind/condo-cli/internal/git"
	kube "github.com/automotiveMastermind/condo-cli/internal/kubernetes"
	"github.com/automotiveMastermind/condo-cli/internal/mongo"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

// ClusterOptions holds the options specific to cluster creation
type ClusterOptions struct {
	Name    string
	Image   string
	Version string
}

const clusterUseStr string = "cluster"

var (
	// cluster options defaults
	clusterOptions = &ClusterOptions{
		Name:    "local",
		Version: "v1.18.19",
		Image:   "kindest/node",
	}

	//go:embed template/cluster/config.yaml
	CLUSTER_CONFIG_FILE_BYTES []byte

	//go:embed template/cluster/git-service.yaml
	CLUSTER_GIT_SERVICE_FILE_BYTES []byte

	//go:embed template/cluster/registry-configmap.yaml
	CLUSTER_CONFIG_MAP_FILE_BYTES []byte

	kubeClient      *kubernetes.Clientset
	clusterRootPath = ""

	/*
		KNOWN ISSUE:
		Reference: https://github.com/spf13/cobra/issues/362

		A command cannot have more than one parent command or else it only
		attaches to the last parent command attached.
		(works like a pointer reference)

		Temp solution:
		Create multiple cluster commands that attach to the different
		parent with different run functions

	*/

	// clusterCmd represents the cluster command specific to the create command
	clusterCreateCmd = &cobra.Command{
		Use:   clusterUseStr,
		Short: "Creates a kubernetes cluster on your local docker instance",
		Long: `Creates a kubernetes cluster on your local docker instance. 
On creation it will use the configuration you have already specified in your <clusters> folder. 
If there is no configuration found then a new one will be generated or pulled from 
a git repo indicated in the config.`,
		Run: func(cmd *cobra.Command, args []string) {
			createCluster()
		},
	}

	// clusterCmd represents the cluster command specific to the stop command
	clusterGetCmd = &cobra.Command{
		Use:   clusterUseStr,
		Short: "Gets the kubernetes cluster on your local docker instance",
		Long:  `Gets the kubernetes cluster on your local docker instance`,
		Run: func(cmd *cobra.Command, args []string) {
			//getCluster()
		},
	}
)

func init() {
	// flags
	clusterCreateCmd.Flags().StringVar(&clusterOptions.Name, "name", clusterOptions.Name, "Sets the name of the cluster")
	clusterCreateCmd.Flags().StringVar(&clusterOptions.Image, "image", clusterOptions.Image, "Sets the image to use for the cluster")
	clusterCreateCmd.Flags().StringVar(&clusterOptions.Version, "version", clusterOptions.Version, "Sets the image version for the cluster")

	//clusterStopCmd.Flags().StringVar(&clusterOptions.Name, "name", clusterOptions.Name, "Sets the name of the cluster")

	// add cluster cmd to create
	createCmd.AddCommand(clusterCreateCmd)

	//add cluster cmd to stop
	//stopCommand.AddCommand(clusterStopCmd)

	// add cluster cmd to get
	getCommand.AddCommand(clusterGetCmd)

	// cluster stop
	stopCmd := cluster.NewCmdStop()
	stopCmd.Flags().StringVar(&clusterOptions.Name, "name", clusterOptions.Name, "Sets the name of the cluster")
	stopCommand.AddCommand(stopCmd)
}

//entry point for "create cluster"
func createCluster() {
	log.Info("Hello! Welcome to condo create cluster!")
	checkExecDependencies()

	if !clusterConfigExists(clusterOptions.Name) {
		createDefaultClusterConfig()
		git.CreateConfig(clusterRootPath, clusterOptions.Name, !configExists("deploy"), !configExists("helm"))
	}

	if isClusterRunning() {
		log.Fatalf("Cluster '%s' is already running", clusterOptions.Name)
	}

	startCluster()
	kubeClientCreate()

	log.Info("Init cluster...")
	createNamespaces()
	createPolicies()
	createIngress()
	git.Run(clusterRootPath)
	docker.Run()
	mongo.Run()
	installSealedSecrets()
	flux.Install(clusterOptions.Name, clusterRootPath)
	log.Infof("Cluster '%s' ready please add your deployments in (%s)", clusterOptions.Name, clusterRootPath+"/deploy")
}

func checkExecDependencies() {
	docker.IsRunning()
	checkExecExists("kind")
	checkExecExists("git")
	checkExecExists("kubectl")
	checkExecExists("helm")
	checkExecExists("docker")
}

func checkExecExists(executable string) {
	path, err := exec.LookPath(executable)
	if err != nil {
		log.Fatalf("'%s' executable not found", executable)
	} else {
		log.Debugf("'%s' executable found at '%s'", executable, path)
	}
}

func isClusterRunning() bool {
	cmd := exec.Command("kind", "get", "clusters")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Create cluster failed with %s", err)
	}

	if !(strings.TrimSpace(string(out)) == "No kind clusters found.") {
		log.Fatalf("Only one cluster instance is allowed. Please \"stop\" the previous cluster")
	}

	clusters := bytes.Split(out, []byte("\n"))

	for _, s := range clusters {
		if string(s) == clusterOptions.Name {
			return true
		}
	}
	return false
}

func clusterConfigExists(name string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("can't find users home directory")
	}

	clusterRootPath = filepath.Join(home, ".am", "clusters", name)

	if _, err := os.Stat(clusterRootPath); os.IsNotExist(err) {
		log.Infof("Creating new directory for cluster at '%s'", clusterRootPath)
		err := os.MkdirAll(clusterRootPath, 0755)
		check(err)
		return false
	} else {
		log.Infof("Cluster config already exists; will use from directory (%s)", clusterRootPath)
	}
	return true
}

func configExists(folder string) bool {
	path := filepath.Join(clusterRootPath, folder)
	_, err := os.Stat(path)

	return !os.IsNotExist(err)
}

// create cluster default config files from internal binary
func createDefaultClusterConfig() {
	path := filepath.Join(clusterRootPath, "cluster")
	err := os.Mkdir(path, 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %s", err)
	}

	// clusterRootPath already set by previous method
	log.Info("Creating cluster configuration")

	// write main config
	errMainConfig := ioutil.WriteFile(filepath.Join(path, "config.yaml"), CLUSTER_CONFIG_FILE_BYTES, 0644)
	if errMainConfig != nil {
		log.Fatalf("Embedded file \"config.yaml\" failed to write to directory")
	}

	// write git service config
	errGitService := ioutil.WriteFile(filepath.Join(path, "git-service.yaml"), CLUSTER_GIT_SERVICE_FILE_BYTES, 0644)
	if errGitService != nil {
		log.Fatalf("Embedded file \"git-service.yaml\" failed to write to directory")
	}

	//write registry map config
	errConfigMap := ioutil.WriteFile(filepath.Join(path, "registry-configmap.yaml"), CLUSTER_CONFIG_MAP_FILE_BYTES, 0644)
	if errConfigMap != nil {
		log.Fatalf("Embedded file \"registry-configmap.yaml\" failed to write to directory")
	}

	log.Info("Cluster configurations created")
}

func startCluster() {
	// install via kind
	imageFlag := fmt.Sprintf("--image=%s:%s", clusterOptions.Image, clusterOptions.Version)
	nameFlag := fmt.Sprintf("--name=%s", clusterOptions.Name)
	configFlag := fmt.Sprintf("--config=%s", filepath.Join(clusterRootPath, "cluster", "config.yaml"))

	cmd := exec.Command("kind", "create", "cluster", imageFlag, nameFlag, configFlag)
	err := console.Start(cmd)

	if err != nil {
		log.Fatalf("Failed to start command kind %s\n", err)
	}
}

func kubeClientCreate() {
	log.Infof("Ensuring kubectl context change to kind-%s", clusterOptions.Name)
	kubeClient = kube.BuildClient("kind-" + clusterOptions.Name)
}

func createIngress() {
	log.Info("Creating ingress...")
	deploymentURI := "https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v0.46.0/deploy/static/provider/kind/deploy.yaml"

	cmd := exec.Command("kubectl", "apply", "-f", deploymentURI)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to create ingress: %v", err)
	}
}

func createNamespaces() {
	log.Info("Creating namespaces...")
	path := filepath.Join(clusterRootPath, "helm", ".cluster", "namespaces")
	cmd := exec.Command("kubectl", "apply", "-f", path)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to create namespaces: %v", err)
	}
}

func createPolicies() {
	log.Info("Creating policies...")
	path := filepath.Join(clusterRootPath, "helm", ".cluster")
	cmd := exec.Command("kubectl", "apply", "-f", path)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to create policies: %v", err)
	}
}

func installSealedSecrets() {
	log.Info("Starting sealed secrets...")
	cmd := exec.Command(
		"helm",
		"upgrade",
		"sealed-secrets-controller",
		filepath.Join(clusterRootPath, "helm", "sealed-secrets"),
		"--install",
		"--wait",
		"--namespace=kube-system",
		"--values="+filepath.Join(clusterRootPath, "helm", ".values", "sealed-secrets.yaml"),
	)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to start sealed secrets: %v", err)
	}

	// reload secret if exists
	secretsPath := filepath.Join(clusterRootPath, ".secrets", "sealed-secrets.yaml")
	if _, err = os.Stat(secretsPath); os.IsExist(err) {
		secret, err := ioutil.ReadFile(secretsPath)
		check(err)
		cmd = exec.Command(
			"kubectl",
			"apply",
			"--overwrite=true",
			"-f",
			"-"+string(secret),
		)

		err = cmd.Run()
		if err != nil {
			log.Fatalf("failed to overwrite sealed secret from saved yaml: %v", err)
		}
	}

	// save secret to cluster config
	cmd = exec.Command(
		"kubectl",
		"get",
		"secret",
		"--namespace=kube-system",
		"-l",
		"sealedsecrets.bitnami.com/sealed-secrets-key",
		"--output=yaml",
	)
	secret, err := cmd.Output()
	check(err)

	err = ioutil.WriteFile(secretsPath, secret, 0644)
	check(err)
}

func clusterExistCheck() {

	log.Info("Checking that cluster \"" + clusterOptions.Name + "\" exists...")

	//TO-DO check to see if the cluster exists

	out, err := exec.Command("kind", "get", "clusters").Output()

	if err != nil {
		log.Fatalf("Unknown kind error  %v", err)
	}

	outputStr := string(out)
	outputArray := strings.Fields(outputStr)

	if contains(outputArray, clusterOptions.Name) {
		log.Info("Cluster detected")
	} else {
		log.Fatal("Cluster not found, aborting operation...")
	}
}

//check if a string equivalent exists in a string array
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

//removes cluster nodes using kind. Use '--name [clusterName]' to specify the cluster name if not default
func removeClusterNodes() {
	log.Info("Removing cluster \"" + clusterOptions.Name + "\" from docker...")

	nameFlag := fmt.Sprintf("--name=%s", clusterOptions.Name)

	cmd := exec.Command("kind", "delete", "cluster", nameFlag)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("Failed to remove cluster:  %v", err)
	}
	log.Info("cluster \"" + clusterOptions.Name + "\" removed from docker")
}

func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
	}
}
