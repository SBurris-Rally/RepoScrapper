package data

import (
    "fmt"
    "os"
    "os/exec"
    "time"

    utils "Sburris/reposcrapper/src/utils"
)

type Repository struct {
	ID int
	NodeID string
	Name string
	FullName string
	IsPrivate bool
	IsFork	bool
	HTMLURL string
	Description string
	CreatedAt time.Time
	UpdatedAt time.Time
	PushedAt time.Time
	DefaultBranch string
	// TODO: is this still needed if clone is local?
	//Files []*RepositoryFile
	WasTreeTruncated bool
	SSHURL string

	// From Repoman files
	Status string
	Team string
	JiraProject string
	IsLive string
}

func (repo *Repository) Update() {
	targetDir := fmt.Sprintf("%s/%s", DefaultReposLocation, repo.Name)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		repo.Clone(targetDir)
	} else {
		repo.Pull(targetDir)
	}
}

func (repo *Repository) Delete() {
	targetDir := fmt.Sprintf("%s/%s", DefaultReposLocation, repo.Name)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		utils.Log(fmt.Sprintf("Repo '%s' does not exist at location: %s", repo.Name, targetDir))
	} else {
		err := os.RemoveAll(targetDir)
		utils.CheckError(err, fmt.Sprintf("Removing Repo '%s' at location: %s", repo.Name, targetDir))
	}
}

func (repo *Repository) Clone(targetDir string) {
	//Log(fmt.Sprintf("Cloning Repo: %s", repo.Name ))

	// command: git clone <repo> <location>
	fmt.Printf("Cloning (%s): git clone %s %s\n", repo.Name, repo.SSHURL, targetDir)
	cmd := exec.Command("git", "clone", repo.SSHURL, targetDir)
	//fmt.Println(cmd)
		
	err := cmd.Run()
	utils.CheckError(err, "Failed cloning repo")

}

func (repo *Repository) Pull(targetDir string) {
	//Log(fmt.Sprintf("Pulling Repo: %s", repo.Name ))

	// command: git clone <repo> <location>
	fmt.Printf("Pulling (%s): git -C %s pull\n", repo.Name, targetDir)
	cmd := exec.Command("git", "-C", targetDir, "pull")
	//fmt.Println(cmd)

	err := cmd.Run()
	utils.CheckError(err, "Failed to pull repo")
}