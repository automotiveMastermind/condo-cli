package git

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	DEPLOY_CONFIG_GIT_REPO        string
	DEPLOY_CONFIG_GIT_REPO_BRANCH string

	HELM_CONFIG_GIT_REPO        string
	HELM_CONFIG_GIT_REPO_BRANCH string
}

//region public functions
func CreateConfig(clusterRootPath string, clusterName string, withDeploy bool, withHelm bool) {
	configuration := loadConfig()
	if withDeploy {
		getGitRepo(configuration.DEPLOY_CONFIG_GIT_REPO, "deploy", configuration.DEPLOY_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
	}

	if withHelm {
		getGitRepo(configuration.HELM_CONFIG_GIT_REPO, "helm", configuration.HELM_CONFIG_GIT_REPO_BRANCH, clusterRootPath, clusterName)
	}
}

//endregion public functions

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

	workingFolderUri := filepath.Join(clusterRootPath, folderName)

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
	commandRmGit.Dir = filepath.Join(clusterRootPath, "tmp", folderName)
	errRmGit := commandRmGit.Run()
	if errRmGit != nil {
		log.Fatalf("Error removing git folder at /tmp/"+folderName+". %s", errRmGit)
	}

	errMove := os.Rename(commandRmGit.Dir, filepath.Join(clusterRootPath, folderName))

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

	errRmTmp := os.RemoveAll(filepath.Join(clusterRootPath, "tmp"))
	if errRmTmp != nil {
		log.Fatalf("Failed to remove tmp folder. %s", errRmTmp)
	}
}

func check(e error) {
	if e != nil {
		log.Fatalf("%v", e)
	}
}
