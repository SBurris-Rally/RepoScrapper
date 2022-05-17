package cmd

import (
    "encoding/csv"
    "os"

	"github.com/spf13/cobra"
	//"Sburris/reposcrapper/src/data"
	//utils "Sburris/reposcrapper/src/utils"
	data "Sburris/reposcrapper/src/data"
	utils "Sburris/reposcrapper/src/utils"
)

var (
    // Used for command line flags
    filePattern     string
    ignorePattern   string
    namespaces      []string
    repositories    []string
    policyLocation  string
    repositoryLocation string
)

var conftestCmd = &cobra.Command {
    Use: "conftest",
    Short: "Working with conftest",
    RunE: func(cmd *cobra.Command, args []string) error {
        cmd.Help()
        return nil
    },
}

func init() {
    // filePattern     string
    // ignorePattern   string
    // namespaces      string // Comma seperated
    // repositories    string // Comma seperated

    conftestCmd.PersistentFlags().StringVarP(&filePattern, "file-pattern", "f", `^(dockerfile|(.*(\.dockerfile|\.yml|\.yaml|\.tf|\.tfvars|\.json)))$`, "regex for selected filenames")
    conftestCmd.PersistentFlags().StringVarP(&ignorePattern, "ignore-pattern", "i", `(\/allure-results\/)|(\/mock-data\/)|(\/.{0,3}test.{0,3}\/)|(\/node_modules\/)`, "regex for filepaths to ignore")
    conftestCmd.PersistentFlags().StringArrayVarP(&namespaces, "namespaces", "n", []string{}, "namespaces to use")
    conftestCmd.PersistentFlags().StringArrayVarP(&repositories, "repositories", "r", []string{}, "repositories to scan")
    conftestCmd.PersistentFlags().StringVarP(&policyLocation, "policy-location", "p", "~/code/conftest-rego-policies/policies/nonblocking", "location of conftest policies")
    conftestCmd.PersistentFlags().StringVarP(&repositoryLocation, "repository-location", "l", "~/.reposcrapper/repos", "base location of repositories to scan")

    conftestCmd.AddCommand(scanCmd)
    rootCmd.AddCommand(conftestCmd)
}

var scanCmd = &cobra.Command {
    Use: "scan",
    Short: "Scan selected repos with the specified conftest namespaces",
    RunE: func(cmd *cobra.Command, args []string) error {
        conftestService := data.CreateConftestService(
            policyLocation,
            filePattern,
            ignorePattern,
            namespaces,
            repositories,
            repositoryLocation,
        )

        results := conftestService.Scan()

        file, err := os.Create("output.csv")
        utils.CheckError(err, "Opening output.csv")
        defer file.Close()

        csvWriter := csv.NewWriter(file)

        headerRow := []string{"Repo", "Team", "Jira", "Status", "IsLive", "Severity", "Namespace", "File", "Message"}

        csvWriter.Write(headerRow)

        for _, result := range results {
            for _, finding := range *result.Results {
                row :=[]string{
                    result.Repo,
                    result.Team,
                    result.Jira,
                    result.Status,
                    result.IsLive,
                    finding.Level,
                    finding.Namespace,
                    finding.Filepath,
                    finding.Message,
                }
                csvWriter.Write(row)
            }
        }
        csvWriter.Flush()

        return nil
    },
}
