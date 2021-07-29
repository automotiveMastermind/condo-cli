/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/pem"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"path/filepath"

	"github.com/creack/pty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

var FPS string = string(filepath.Separator)

//go:embed template/cluster/config.yaml
var CLUSTER_CONFIG_FILE_BYTES []byte

//go:embed template/cluster/git-service.yaml
var CLUSTER_GIT_SERVICE_FILE_BYTES []byte

//go:embed template/cluster/registry-configmap.yaml
var CLUSTER_CONFIG_MAP_FILE_BYTES []byte

// ClusterOptions holds the options specific to cluster creation
type ClusterOptions struct {
	Name    string
	Image   string
	Version string
}

var (
	// cluster options defaults
	clusterOptions = &ClusterOptions{
		Name:    "local",
		Version: "v1.16.15",
		Image:   "kindest/node",
	}

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

	clusterUseStr string = "cluster"

	// clusterCmd represents the cluster command specific to the create command
	clusterCreateCmd = &cobra.Command{
		Use:   clusterUseStr,
		Short: "Creates a kube cluster on your local docker instance",
		Long:  ``,

		Run: func(cmd *cobra.Command, args []string) {

			cluster()

		},
	}
	// clusterCmd represents the cluster command specific to the destroy command
	clusterDestroyCmd = &cobra.Command{
		Use:   clusterUseStr,
		Short: "Removes a kube cluster on your local docker instance",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {

			destroyCluster()

		},
	}
)

func init() {
	// flags
	clusterCreateCmd.Flags().StringVar(&clusterOptions.Name, "name", clusterOptions.Name, "Sets the name of the cluster")
	clusterCreateCmd.Flags().StringVar(&clusterOptions.Image, "image", clusterOptions.Image, "Sets the image to use for the cluster")
	clusterCreateCmd.Flags().StringVar(&clusterOptions.Version, "version", clusterOptions.Version, "Sets the image version for the cluster")

	clusterDestroyCmd.Flags().StringVar(&clusterOptions.Name, "name", clusterOptions.Name, "Sets the name of the cluster")

	// add cluster cmd to create
	createCmd.AddCommand(clusterCreateCmd)

	//add cluster cmd to destroy
	destroyCommand.AddCommand(clusterDestroyCmd)

}

func destroyCluster() {
	dockerCheck()
	clusterExistCheck()
	removeGitServerDockerContainer()
	removeDockerRegistryDockerContainer()
	removeClusterNodes()

	log.Info("Cluster destroyed")

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

//removes 'git-server' container from the system's instance of docker
func removeGitServerDockerContainer() {
	log.Info("Removing container git-server from docker")

	//stop the git-server container
	dockerStopCmd := exec.Command(
		"docker",
		"stop",
		"git-server",
	)
	var dockerStopErr error
	dockerStopErr = dockerStopCmd.Run()
	if dockerStopErr != nil {

		log.Infof("Docker container \"git-server\" failed to stop:  %v", dockerStopErr)
	}

	//remove the git-server container
	dockerRemoveCmd := exec.Command(
		"docker",
		"rm",
		"git-server",
	)
	var dockerRemoveErr error
	dockerRemoveErr = dockerRemoveCmd.Run()
	if dockerRemoveErr != nil {

		log.Infof("Docker container \"git-server\" failed to be removed:  %v", dockerRemoveErr)
	}

	log.Info("git-server removed from docker")

}

//removes cluster nodes using kind. Use '--name [clusterName]' to specify the cluster name if not default
func removeClusterNodes() {
	log.Info("Removing cluster \"" + clusterOptions.Name + "\" from docker...")

	nameFlag := fmt.Sprintf("--name=%s", clusterOptions.Name)

	cmd := exec.Command("kind", "delete", "cluster", nameFlag)

	var err error
	err = cmd.Run()
	if err != nil {
		log.Fatalf("Failed to remove cluster:  %v", err)
	}
	log.Info("cluster \"" + clusterOptions.Name + "\" removed from docker")

}

//entry point for "create cluster"
func cluster() {
	log.Info("Hello! Welcome to condo create cluster!")
	dockerCheck()
	checkExecDependencies()
	if !clusterConfigExists(clusterOptions.Name) {
		createDefaultClusterConfig()
		CreateAuxilaryConfig(clusterRootPath, clusterOptions.Name)

		//copy existing sealedSecret || TO BE REMOVED TO REPLACED WITH GENERATED SEALED SECRET
		copySealedSecret()
	} else if !clusterAuxiliaryConfigExists(clusterOptions.Name, "deploy") {
		CreateAuxilaryConfigDeployOnly(clusterRootPath, clusterOptions.Name)

	} else if !clusterAuxiliaryConfigExists(clusterOptions.Name, "helm") {
		CreateAuxilaryConfigHelmOnly(clusterRootPath, clusterOptions.Name)

	}

	if isClusterRunning() {
		log.Fatalf("Cluster '%s' is already running", clusterOptions.Name)
	}
	createCluster()

	log.Info("Init cluster...")
	createNamespaces()
	createPolicies()
	createIngress()
	createRSAKey()
	installGitServer()
	configGitInCluster()
	installDockerRegistry()
	installSealedSecrets()
	installFluxSecrets()
	installFlux()
	installFluxHelmOperator()

	log.Infof("Cluster '%s' ready please add your deployments in (%s)", clusterOptions.Name, clusterRootPath+"/deploy")
}

func checkExecDependencies() {
	checkExecExists("kind")
	checkExecExists("git")
	checkExecExists("kubectl")
	checkExecExists("helm")
	checkExecExists("docker")
}

func checkExecExists(executable string) {
	path, err := exec.LookPath(executable)
	if err != nil {
		log.Warningf("'%s' executable not found\n", executable)
		panic(err)
	} else {
		log.Debugf("'%s' executable found at '%s'\n", executable, path)
	}
}

func isClusterRunning() bool {
	cmd := exec.Command("kind", "get", "clusters")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Create cluster failed with %s\n", err)
	}
	if !(strings.TrimSpace(string(out)) == "No kind clusters found.") {
		log.Fatalf("Only one cluster instance is allowed. Please \"destroy\" the previous cluster")
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
		log.Fatal("can't find users home directory\n")
	}

	clusterRootPath = fmt.Sprintf(home+FPS+".am"+FPS+"clusters"+FPS+"%s", name)

	if _, err := os.Stat(clusterRootPath); os.IsNotExist(err) {
		log.Infof("Creating new directory for cluster at '%s'\n", clusterRootPath)
		err := os.MkdirAll(clusterRootPath, 0755)
		check(err)
		return false
	} else {

		log.Infof("Cluster config already exists; will use from directory (%s)", clusterRootPath)

	}
	return true
}

func clusterAuxiliaryConfigExists(name string, folder string) bool {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("can't find users home directory\n")
	}

	clusterRootPath = fmt.Sprintf(home+FPS+".am"+FPS+"clusters"+FPS+"%s", name)

	if _, err := os.Stat(clusterRootPath + FPS + folder); os.IsNotExist(err) {

		return false
	}
	return true
}

// create cluster default config files from internal binary
func createDefaultClusterConfig() {

	err := os.Mkdir(clusterRootPath+FPS+"cluster", 0755)
	if err != nil {
		log.Fatalf("Failed to create directory: %s", err)
	}

	//clusterRootPath already set by previous method
	log.Info("Creating cluster configuration")

	//write main config
	errMainConfig := ioutil.WriteFile(clusterRootPath+FPS+"cluster"+FPS+"config.yaml", CLUSTER_CONFIG_FILE_BYTES, 0644)
	if errMainConfig != nil {
		log.Fatalf("Embeded file \"config.yaml\" failed to write to directory")
	}

	//write git service config
	errGitService := ioutil.WriteFile(clusterRootPath+FPS+"cluster"+FPS+"git-service.yaml", CLUSTER_GIT_SERVICE_FILE_BYTES, 0644)
	if errGitService != nil {
		log.Fatalf("Embeded file \"git-service.yaml\" failed to write to directory")
	}

	//write registry map config
	errConfigMap := ioutil.WriteFile(clusterRootPath+FPS+"cluster"+FPS+"registry-configmap.yaml", CLUSTER_CONFIG_MAP_FILE_BYTES, 0644)
	if errConfigMap != nil {
		log.Fatalf("Embeded file \"registry-configmap.yaml\" failed to write to directory")
	}

	log.Info("Cluster configurations created")

}

func createCluster() {
	// install via kind
	imageFlag := fmt.Sprintf("--image=%s:%s", clusterOptions.Image, clusterOptions.Version)
	nameFlag := fmt.Sprintf("--name=%s", clusterOptions.Name)
	configFlag := fmt.Sprintf("--config=%s/cluster/config.yaml", clusterRootPath)

	cmd := exec.Command("kind", "create", "cluster", imageFlag, nameFlag, configFlag)
	var err error
	if runtime.GOOS == "windows" {
		var stdoutBuf, stderrBuf bytes.Buffer
		cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
		cmd.Stderr = io.MultiWriter(os.Stderr, &stderrBuf)
		err = cmd.Run()
	} else {
		// run in faketty to get better looking output
		f, _ := pty.Start(cmd)
		io.Copy(os.Stdout, f)
	}

	if err != nil {
		log.Fatalf("Failed to start command kind %s\n", err)
	}

	log.Infof("Ensuring kubectl context change to kind-%s", clusterOptions.Name)
	cmd = exec.Command(
		"kubectl",
		"cluster-info",
		"--context",
		"kind-"+clusterOptions.Name,
	)

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Could not change cluster context to %s", "kind-"+clusterOptions.Name)
	}
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
	//TO-DO remove
	log.Info("kubectl apply -f " + clusterRootPath + FPS + "helm" + FPS + ".cluster" + FPS + "namespaces")

	cmd := exec.Command("kubectl", "apply", "-f", clusterRootPath+FPS+"helm"+FPS+".cluster"+FPS+"namespaces")

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to create namespaces: %v", err)
	}
}

func createPolicies() {
	log.Info("Creating policies...")
	cmd := exec.Command("kubectl", "apply", "-f", clusterRootPath+FPS+"helm"+FPS+".cluster")

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to create policies: %v", err)
	}
}

func createRSAKey() {
	sshKeyPath := clusterRootPath + FPS + ".ssh"
	privateKeyPath := sshKeyPath + FPS + "identity"
	publicKeyPath := sshKeyPath + FPS + "identity.pub"

	log.Debugf("ssh-path: %s \nprivate-key: %s \npublic-key: %s \n", sshKeyPath, privateKeyPath, publicKeyPath)

	_, err := os.Stat(privateKeyPath)
	_, err2 := os.Stat(publicKeyPath)
	if err == nil || err2 == nil {
		log.Info("RSA keys already exist...")
		return
	}

	log.Info("Creating RSA keys...")
	if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
		err = os.Mkdir(sshKeyPath, 0755)
		check(err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	check(err)

	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)
	check(err)

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	err = ioutil.WriteFile(publicKeyPath, pubKeyBytes, 0600)
	check(err)

	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	err = ioutil.WriteFile(privateKeyPath, privatePEM, 0600)
	check(err)
}

func installGitServer() {
	log.Info("Starting git server...")

	cmd := exec.Command(
		"docker",
		"run",
		"-d",
		"-p2222:22",
		"--pull=missing",
		"--name=git-server",
		"-v"+clusterRootPath+":/git-server/repos",
		"-v"+clusterRootPath+"/.ssh:/git-server/keys",
		"jkarlos/git-server-docker",
	)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to start git server: %v", err)
	}

	// connect to kind network
	cmd = exec.Command(
		"docker",
		"network",
		"connect",
		"kind",
		"git-server",
	)

	err = cmd.Run()
	check(err)
}

func configGitInCluster() {
	log.Debug("attach git-server to kind network")

	// find docker ip for git server
	format := `{{range .Containers}}{{if eq .Name "git-server"}}{{.IPv4Address}}{{end}}{{- end}}`
	cmd := exec.Command(
		"docker",
		"network",
		"inspect",
		"kind",
		"--format",
		format,
	)

	ip, err := cmd.Output()
	check(err)

	log.Debug("get template")

	// apply ip to template
	tpl, err := template.ParseFiles(clusterRootPath + "/cluster/git-service.yaml")
	check(err)

	log.Debug("template parsed")

	var doc bytes.Buffer
	tpl.Execute(&doc, string(bytes.Split(ip, []byte("/"))[0]))
	check(err)

	log.Debug("template injected")

	// attach ip to endpoint and service
	echo := exec.Command("echo", doc.String())
	cmd = exec.Command(
		"kubectl",
		"apply",
		"--overwrite=true",
		"-f",
		"-",
	)

	pipe, _ := echo.StdoutPipe()
	defer pipe.Close()

	cmd.Stdin = pipe
	echo.Start()

	res, _ := cmd.Output()

	log.Debug(string(res))

	log.Debug("service applied")

}

func installSealedSecrets() {
	log.Info("Starting sealed secrets...")
	cmd := exec.Command(
		"helm",
		"upgrade",
		"sealed-secrets-controller",
		clusterRootPath+FPS+"helm"+FPS+"sealed-secrets",
		"--install",
		"--wait",
		"--namespace=kube-system",
		"--values="+clusterRootPath+FPS+"helm"+FPS+".values"+FPS+"sealed-secrets.yaml",
	)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to start sealed secrets: %v", err)
	}

	// reload secret if exists
	secretsPath := clusterRootPath + FPS + ".secrets" + FPS + "sealed-secrets.yaml"
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

func installFluxSecrets() {
	log.Info("Creating flux secrets")

	cmd := exec.Command(
		"kubectl",
		"create",
		"secret",
		"generic",
		"flux-git-deploy",
		"--from-file="+clusterRootPath+FPS+".ssh"+FPS+"identity",
		"--namespace=weave",
	)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to create flux secret: %v", err)
	}
}

func installFlux() {
	log.Info("Starting flux...")
	cmd := exec.Command(
		"helm",
		"upgrade",
		"flux",
		clusterRootPath+"/helm/fluxcd/flux",
		"--install",
		"--wait",
		"--namespace=weave",
		"--values="+clusterRootPath+FPS+"helm"+FPS+".values"+FPS+"flux.yaml",
		"--set=git.branch="+clusterOptions.Name,
		"--set=git.label=flux-"+clusterOptions.Name,
	)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to start flux: %v", err)
	}
}

func installFluxHelmOperator() {
	log.Info("Starting flux helm operator...")
	cmd := exec.Command(
		"helm",
		"upgrade",
		"flux-helm-operator",
		clusterRootPath+FPS+"helm"+FPS+"fluxcd"+FPS+"helm-operator",
		"--install",
		"--wait",
		"--namespace=weave",
		"--values="+clusterRootPath+FPS+"helm"+FPS+".values"+FPS+"helm-operator.yaml",
		"--set=helm.versions=v3",
	)

	err := cmd.Run()
	if err != nil {
		log.Fatal("failed to start flux helm operator: %v", err)
	}
}

//TO_DO remove and replace with generated secret
func copySealedSecret() {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("can't find users home directory\n")
	}

	errMk := os.MkdirAll(clusterRootPath+FPS+".secrets"+FPS, 0755)
	check(errMk)

	commandExec := exec.Command("cp", home+FPS+"sealed-secrets.yaml", clusterRootPath+FPS+".secrets"+FPS)
	//clusterRootPath already set by a preceeding method

	copyErr := commandExec.Run()
	if copyErr != nil {

		log.Fatalf("Sealed Secret Copy error:  %s , home:"+home+", clusterRootPath: "+clusterRootPath, copyErr)
	}

	permChangeErr := os.Chmod(clusterRootPath+FPS+".secrets"+FPS+"sealed-secrets.yaml", 0755)
	check(permChangeErr)

}

//check if docker engine is running
func dockerCheck() {

	log.Info("Looking for docker instance...")
	cmd := exec.Command("docker", "ps")

	var err error
	err = cmd.Run()
	if err != nil {
		log.Fatalf("No docker instance found:  %v", err)
	}

	log.Info("Docker instance found")

}

func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
	}
}
