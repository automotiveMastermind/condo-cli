package git

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"html/template"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/automotiveMastermind/condo-cli/internal/docker"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var GIT_INSTANCE_NAME string = "git-server"

func Run(clusterRootPath string) {
	createRSAKey(clusterRootPath)
	startGitServer(clusterRootPath)
	configGitInCluster(clusterRootPath)
}

//removes 'git-server' container from the system's instance of docker
func Stop() {
	log.Info("Removing container git-server from docker")

	//stop the git-server container
	dockerStopCmd := exec.Command(
		"docker",
		"stop",
		GIT_INSTANCE_NAME,
	)

	dockerStopErr := dockerStopCmd.Run()
	if dockerStopErr != nil {

		log.Infof("Docker container \"%s\" failed to stop:  %v", GIT_INSTANCE_NAME, dockerStopErr)
	}

	//remove the git-server container
	dockerRemoveCmd := exec.Command(
		"docker",
		"rm",
		GIT_INSTANCE_NAME,
	)

	dockerRemoveErr := dockerRemoveCmd.Run()
	if dockerRemoveErr != nil {
		log.Infof("Docker container \"%s\" failed to be removed:  %v", GIT_INSTANCE_NAME, dockerRemoveErr)
	}

	log.Info("git-server removed from docker")
}

func createRSAKey(clusterRootPath string) {
	sshKeyPath := filepath.Join(clusterRootPath, ".ssh")
	privateKeyPath := filepath.Join(sshKeyPath, "identity")
	publicKeyPath := filepath.Join(sshKeyPath, "identity.pub")

	log.Tracef("ssh-path: %s \nprivate-key: %s \npublic-key: %s \n", sshKeyPath, privateKeyPath, publicKeyPath)

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

func startGitServer(clusterRootPath string) {
	log.Info("Starting git server...")

	if docker.IsImageRunning(GIT_INSTANCE_NAME) {
		log.Info(GIT_INSTANCE_NAME + " is already running, skipping " + GIT_INSTANCE_NAME + " creation.")
		return
	}

	cmd := exec.Command(
		"docker",
		"run",
		"-d",
		"-p2222:22",
		"--pull=missing",
		"--name="+GIT_INSTANCE_NAME,
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

func configGitInCluster(clusterRootPath string) {
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

	// apply ip to template
	tpl, err := template.ParseFiles(clusterRootPath + "/cluster/git-service.yaml")
	check(err)

	var doc bytes.Buffer
	tpl.Execute(&doc, string(bytes.Split(ip, []byte("/"))[0]))
	check(err)

	log.Trace("Parsed Template (git-service.yaml):")
	log.Trace(doc.String())

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

	log.Trace(string(res))
	log.Debug("service applied")
}
