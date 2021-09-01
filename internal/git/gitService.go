package git

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

//gets the native OS' filepath separator character, stores it in 'FPS' to use when defining file paths
var FPS string = string(filepath.Separator)

//region public functions
func CreateAuxilaryConfig(clusterRootPath string, clusterName string) {
	configuration := loadConfig()
	getGitRepo(configuration.DEPLOY_CONFIG_GIT_REPO, "deploy", configuration.DEPLOY_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
	getGitRepo(configuration.HELM_CONFIG_GIT_REPO, "helm", configuration.HELM_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
}

func CreateAuxilaryConfigDeployOnly(clusterRootPath string, clusterName string) {
	configuration := loadConfig()
	getGitRepo(configuration.DEPLOY_CONFIG_GIT_REPO, "deploy", configuration.DEPLOY_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
}

func CreateAuxilaryConfigHelmOnly(clusterRootPath string, clusterName string) {
	configuration := loadConfig()
	getGitRepo(configuration.HELM_CONFIG_GIT_REPO, "helm", configuration.HELM_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
}

//endregion public functions

type Configuration struct {
	DEPLOY_CONFIG_GIT_REPO        string
	DEPLOY_CONFIG_GIT_REPO_BRANCH string

	HELM_CONFIG_GIT_REPO        string
	HELM_CONFIG_GIT_REPO_BRANCH string
}

func loadConfig() Configuration {
	file, _ := os.Open("config.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatalf("Failed to load config.json. %s", err)
	}

	return configuration
}

func setUpLocalGitFolder(folderName string, clusterRootPath string, clusterName string) {

	moveCloneIntoLocalRepo(folderName, clusterRootPath)

	workingFolderUri := clusterRootPath + FPS + folderName

	commandExec := exec.Command("git", "init")

	commandExec.Dir = workingFolderUri
	errGitInit := commandExec.Run()
	if errGitInit != nil {
		log.Fatalf("Failed to initialize local git repo "+folderName+". %s", errGitInit)
	}

	commandSwitchBranch := exec.Command("git", "switch", "-c", clusterName)
	commandSwitchBranch.Dir = workingFolderUri
	errSwitchBranch := commandSwitchBranch.Run()
	if errSwitchBranch != nil {
		log.Fatalf("Failed to switch branch of local git repo "+folderName+". %s", errSwitchBranch)
	}

	commandGitAdd := exec.Command("git", "add", "-A")
	commandGitAdd.Dir = workingFolderUri
	errGitAdd := commandGitAdd.Run()
	if errGitAdd != nil {
		log.Fatalf("Failed to initialize local git repo "+folderName+". %s", errGitAdd)
	}

	commandGitCommit := exec.Command("git", "commit", "-m", "\"INIT COMMIT\"")
	commandGitCommit.Dir = workingFolderUri
	errGitCommit := commandGitCommit.Run()
	if errGitCommit != nil {
		log.Fatalf("Failed to initialize local git repo "+folderName+". %s", errGitCommit)
	}

}

func moveCloneIntoLocalRepo(folderName string, clusterRootPath string) {

	commandRmGit := exec.Command("rm", "-rf", ".git")
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

	err := os.MkdirAll(clusterRootPath+"/tmp", 0755)
	check(err)

	commandExec := exec.Command("git", "clone", "--branch", branchName, gitUrl, folderName)
	commandExec.Dir = clusterRootPath + "/tmp"
	out, err := commandExec.CombinedOutput()
	if err != nil {
		log.Fatalf("Failed to clone auxiliary configurations for \""+folderName+"\" into tmp folder. \n Verify Git is installed and Git Credential Manager is configured. (Ref: https://docs.kube.network/tutorials/developer-setup/). \n %s", err)
	}

	log.Infof("%s", out)

	setUpLocalGitFolder(folderName, clusterRootPath, clusterName)

	errRmTmp := os.RemoveAll(clusterRootPath + FPS + "tmp")
	if errRmTmp != nil {
		log.Fatalf("Failed to remove tmp folder. %s", errRmTmp)
	}
}

func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
	}
}
