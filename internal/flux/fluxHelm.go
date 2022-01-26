package flux

import (
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func Install(name, root string) {
	createSecret(root)
	fluxInstall(name, root)
	helmInstall(root)
}

func createSecret(root string) {
	log.Info("Creating flux secrets")

	path := filepath.Join(root, ".ssh", "identity")
	cmd := exec.Command(
		"kubectl",
		"create",
		"secret",
		"generic",
		"flux-git-deploy",
		"--from-file="+path,
		"--namespace=weave",
	)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to create flux secret: %v", err)
	}
}

func fluxInstall(name, root string) {
	log.Info("Starting flux...")

	path := filepath.Join(root, "helm", "fluxcd", "flux")
	values := filepath.Join(root, "helm", ".values", "flux.yaml")
	cmd := exec.Command(
		"helm",
		"upgrade",
		"flux",
		path,
		"--install",
		"--wait",
		"--namespace=weave",
		"--values="+values,
		"--set=git.branch="+name,
		"--set=git.label=flux-"+name,
	)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to start flux: %v", err)
	}
}

func helmInstall(root string) {
	log.Info("Starting flux helm operator...")

	path := filepath.Join(root, "helm", "fluxcd", "helm-operator")
	values := filepath.Join(root, "helm", ".values", "helm-operator.yaml")
	cmd := exec.Command(
		"helm",
		"upgrade",
		"flux-helm-operator",
		path,
		"--install",
		"--wait",
		"--namespace=weave",
		"--values="+values,
		"--set=helm.versions=v3",
	)

	err := cmd.Run()
	if err != nil {
		log.Fatalf("failed to start flux helm operator: %v", err)
	}
}
