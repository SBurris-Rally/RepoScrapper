package cmd

import (	
    "fmt"
    "os"

	"github.com/spf13/cobra"

    "Sburris/reposcrapper/src/data"
    utils "Sburris/reposcrapper/src/utils"
)

const DEFAULT_HOME_FOLDER string = "/Users/stephen.burris/.reposcrapper"

var reposCmd = &cobra.Command {
    Use: "repos",
    Short: "Working with our repositories",
    RunE: func(cmd *cobra.Command, args []string) error {
        cmd.Help()
        return nil
    },
}

func init() {
    reposCmd.AddCommand(updateRepoCmd)
    reposCmd.AddCommand(listOurReposCmd)
    reposCmd.AddCommand(mapToTeamCmd)
    rootCmd.AddCommand(reposCmd)
}

var listOurReposCmd = &cobra.Command {
    Use: "our-repos",
    Short: "List repos belonging to AppSec",
    RunE: func(cmd *cobra.Command, args []string) error {
        myData := data.NewRepoScrapperData()

        outputFilename := fmt.Sprintf("%s%s", DEFAULT_HOME_FOLDER, "/db.raw.json")

        if _, err := os.Stat(outputFilename); os.IsNotExist(err) {
            utils.Log("Cache not found, retrieving Repositories")

            myData.Repos = make(map[string]*data.Repository)

            utils.Log("Finished Processing Repos")
        } else {
            utils.Log("Data Found - loading from cache")
            myData.LoadData(outputFilename)
            utils.Log("Finished")
        }

        lookup := []string{"devsecops", "appsec", "seceng", "security", "secint", "so", "secst"}
        repos := myData.IdentifyReposBasedOnRepomanValue(lookup)

        utils.SaveSliceOfStringsToFile("repos2.csv", repos)

        return nil
    },
}

var updateRepoCmd = &cobra.Command {
    Use: "update",
    Short: "Update local copies of our Github Repositories",
    RunE: func(cmd *cobra.Command, args []string) error {
        myData := data.NewRepoScrapperData()

        myData.UpdateLocalRepos()
        myData.UpdateMetadataFromRepoman()
        //data.IdentifyNoncompliantPullRequests()
        //lookup := []string{"devsecops", "appsec", "seceng", "security", "secint", "so", "secst"}
        //repos := data.IdentifyReposBasedOnRepomanValue(lookup)

        // for _, repo := range repos {
        // 	fmt.Println(repo)
        // }

        //data.CloneAllRepos()
        //data.CreateFileList()
        //data.DisplaySummary()

        myData.SaveRepos(false)


        // for _, r := range allRepos {
        // 	if r.Name == "webdriverio-temporary" {
        // 		temp := make([]Repository, 1)
        // 		temp[0]=r
        // 		GetRepoTrees(&temp)
        // 	}
        // }

        //DownloadMatchedFiles("default\\.rb", *repos)
        //pattern := "(?i)(dockerfile)|(jenkinsfile)|(\\.sh$)|(ignore)|(docker-compose\\.yml$)"
        //pattern := "(?i)docker-compose\\.yml$"
        //pattern := "(?i)package\\.json$"
        //pattern := "nginx\\.conf$"
        //pattern := "\\.tf$"
        //data.DownloadFiles(pattern)
        return nil
    },
}

var mapToTeamCmd = &cobra.Command {
    Use: "map",
    Short: "Map list of repositories to repoman teams",
    RunE: func(cmd *cobra.Command, args []string) error {
        myData := data.NewRepoScrapperData()

        outputFilename := fmt.Sprintf("%s%s", DEFAULT_HOME_FOLDER, "/db.raw.json")
        if _, err := os.Stat(outputFilename); os.IsNotExist(err) {
            fmt.Printf("No cached data found, run update command first.")
            os.Exit(1)
        } else {
            utils.Log("Data Found - loading from cache")
            myData.LoadData(outputFilename)
            utils.Log("Finished")
        }

        // Read list of repositories
        // Split at colon
        // Output CSV containing Repo, Team, Jira, and Status

        inputFile := "springframework_deps.csv"
        myData.MapSnykProjects(inputFile, "springframework_output.csv")

        return nil
    },
}