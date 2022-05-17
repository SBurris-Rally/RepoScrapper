package cmd

import (
    "fmt"

	"github.com/spf13/cobra"

    //"Sburris/reposcrapper/src/data"
    //utils "Sburris/reposcrapper/src/utils"
    data "Sburris/reposcrapper/src/data"
    utils "Sburris/reposcrapper/src/utils"
)

var snykCmd = &cobra.Command {
    Use: "snyk",
    Short: "Working with Snyk",
    RunE: func(cmd *cobra.Command, args []string) error {
        cmd.Help()
        return nil
    },
}

func init() {
    snykCmd.AddCommand(findMissingCmd)
    snykCmd.AddCommand(checkSyncCmd)
    rootCmd.AddCommand(snykCmd)
}

var findMissingCmd = &cobra.Command {
    Use: "missing",
    Short: "Identify which repos are not in Snyk's SCA organization",
    RunE: func(cmd *cobra.Command, args []string) error {
        snykService := data.CreateSnykService(data.SnykAPIBase, data.SnykTokenEnvKey)
        repos := snykService.GetRepos()

        fmt.Println(len(repos))
        // for _, repo := range repos {
        //     fmt.Println(repo)
        // }

        return nil
    },
}

var checkSyncCmd = &cobra.Command {
    Use: "checksync",
    Short: "Check for missing, outdated, and deleted repos",
    RunE: func(cmd *cobra.Command, args []string) error {
        snykService := data.CreateSnykService(data.SnykAPIBase, data.SnykTokenEnvKey)
        snykRepos := snykService.GetRepos()

        githubService := data.CreateGithubDefaultService()
        githubRepos, _ := githubService.GetAllRepos()
        
        githubRepoNames := make([]string, 0, len(githubRepos))
        for _, repo := range githubRepos {
            githubRepoNames = append(githubRepoNames, repo.FullName)
        }

        needToDelete := make([]string, 0)
        for _, repo := range snykRepos {
            if !utils.Contains(githubRepoNames, repo) {
                needToDelete = append(needToDelete, repo)
            }
        }

        needToAdd := make([]string, 0)
        for _, repo := range githubRepoNames {
            if !utils.Contains(snykRepos, repo) {
                needToAdd = append(needToAdd, repo)
            }
        }

        fmt.Println("Need to add:")
        count := len(needToAdd)
        for index, repo := range needToAdd {
            fmt.Printf("(%d/%d) %s\n", index + 1, count, repo)
        }

        fmt.Println("\nNeed to delete:")
        count = len(needToDelete)
        for index, repo := range needToDelete {
            fmt.Printf("(%d/%d) %s\n", index + 1, count, repo)
        }

        return nil
    },
}
