package cmd

import (
	_ "embed"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

var DEPLOY_CONFIG_GIT_REPO string = "https://automotivemastermind@dev.azure.com/automotivemastermind/aM/_git/am.devops.deploy"
var DEPLOY_CONFIG_GIT_REPO_BRANCH string = "local"

var HELM_CONFIG_GIT_REPO string = "https://automotivemastermind@dev.azure.com/automotivemastermind/aM/_git/am.devops.helm"
var HELM_CONFIG_GIT_REPO_BRANCH string = "local"

//var FPS string = string(filepath.Separator)

func setUpLocalGitFolder(folderName string, clusterRootPath string, clusterName string) {

	moveCloneIntoLocalRepo(folderName, clusterRootPath)

	commandExec := exec.Command("git", "init")

	commandExec.Dir = clusterRootPath + FPS + folderName
	errGitInit := commandExec.Run()
	if errGitInit != nil {
		log.Fatalf("Failed to initialize local git repo "+folderName+". %s", errGitInit)
	}

	commandSwitchBranch := exec.Command("git", "switch", "-c", clusterName)
	commandSwitchBranch.Dir = clusterRootPath + FPS + folderName
	errSwitchBranch := commandSwitchBranch.Run()
	if errSwitchBranch != nil {
		log.Fatalf("Failed to switch branch of local git repo "+folderName+". %s", errSwitchBranch)
	}

	commandGitAdd := exec.Command("git", "add", "-A")
	commandGitAdd.Dir = clusterRootPath + FPS + folderName
	errGitAdd := commandGitAdd.Run()
	if errGitAdd != nil {
		log.Fatalf("Failed to initialize local git repo "+folderName+". %s", errGitAdd)
	}

	commandGitCommit := exec.Command("git", "commit", "-m", "\"INIT COMMIT\"")
	commandGitCommit.Dir = clusterRootPath + FPS + folderName
	errGitCommit := commandGitCommit.Run()
	if errGitCommit != nil {
		log.Fatalf("Failed to initialize local git repo "+folderName+". %s", errGitCommit)
	}

}

func cleanTmp(clusterRootPath string) {
	commandRmTmp := exec.Command("rm", "-rf", "tmp")
	commandRmTmp.Dir = clusterRootPath
	errRmTmp := commandRmTmp.Run()
	if errRmTmp != nil {
		log.Fatalf("Failed to remove tmp folder. %s", errRmTmp)
	}

}

func createTmp(clusterRootPath string) {
	err := os.MkdirAll(clusterRootPath+"/tmp", 0755)
	check(err)
}

func CreateAuxilaryConfig(clusterRootPath string, clusterName string) {
	createTmp(clusterRootPath)
	getGitRepo(DEPLOY_CONFIG_GIT_REPO, "deploy", DEPLOY_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
	getGitRepo(HELM_CONFIG_GIT_REPO, "helm", HELM_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
	cleanTmp(clusterRootPath)
}

func CreateAuxilaryConfigDeployOnly(clusterRootPath string, clusterName string) {
	createTmp(clusterRootPath)
	getGitRepo(DEPLOY_CONFIG_GIT_REPO, "deploy", DEPLOY_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
	cleanTmp(clusterRootPath)
}

func CreateAuxilaryConfigHelmOnly(clusterRootPath string, clusterName string) {
	createTmp(clusterRootPath)
	getGitRepo(HELM_CONFIG_GIT_REPO, "helm", HELM_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
	cleanTmp(clusterRootPath)
}

func moveCloneIntoLocalRepo(folderName string, clusterRootPath string) {

	commandRmGit := exec.Command("rm", "-rf", ".git")
	//clusterRootPath already set by a preceeding method
	commandRmGit.Dir = clusterRootPath + FPS + "tmp" + FPS + folderName
	errRmGit := commandRmGit.Run()
	if errRmGit != nil {
		log.Fatalf("Error removing git folder at /tmp/"+folderName+". %s", errRmGit)
	}

	errMove := os.Rename(clusterRootPath+FPS+"tmp"+FPS+folderName, clusterRootPath+FPS+folderName)

	if errMove != nil {
		log.Fatalf("Failed to move files. %v", errMove)
	}

}

func getGitRepo(gitUrl string, folderName string, branchName string, clusterRootPath string, clusterName string) {

	commandExec := exec.Command("git", "clone", "--branch", branchName, gitUrl, folderName)
	//clusterRootPath already set by a preceeding method
	commandExec.Dir = clusterRootPath + "/tmp"
	out, err := commandExec.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to clone auxiliary configurations for \""+folderName+"\" into tmp folder. \n Verify Git is installed and Git Credential Manager is configured. (Ref: https://docs.kube.network/tutorials/developer-setup/). \n %s", err)
	}

	log.Infof("%s", out)

	setUpLocalGitFolder(folderName, clusterRootPath, clusterName)
}
