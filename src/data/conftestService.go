package data

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"regexp"
	"strings"

    "path/filepath"
    "os/user"

	utils "Sburris/reposcrapper/src/utils"
)

type ConftestService struct {
    PoliciesLocation string
    IgnorePattern string
    TargetFilePattern string
    Namespaces []string
    Repositories []string
    BaseRepositoryLocation string
}

func CreateConftestService(policyLocation string, filePattern string, ignorePattern string, namespaces []string, repositories []string, baseRepositoryLocation string) *ConftestService {
    conftestService := ConftestService {
        PoliciesLocation: policyLocation,
        TargetFilePattern: filePattern,
        IgnorePattern: ignorePattern,
        Namespaces: namespaces,
        Repositories: repositories,
        BaseRepositoryLocation: baseRepositoryLocation,
    }

    usr, _ := user.Current()
    dir := usr.HomeDir

    if strings.HasPrefix(conftestService.PoliciesLocation, "~/") {
        conftestService.PoliciesLocation = filepath.Join(dir, conftestService.PoliciesLocation[2:])
    }

    if strings.HasPrefix(conftestService.BaseRepositoryLocation, "~/") {
        conftestService.BaseRepositoryLocation = filepath.Join(dir, conftestService.BaseRepositoryLocation[2:])
    }

    return &conftestService
}

type ConftestFullScanResults struct {
    Repo string
    Team string
    Jira string
    Status string
    IsLive string
    Results *[]ConftestResult
}

func (conftest *ConftestService) Scan() []ConftestFullScanResults {
    // Scan target repo
    // What namespaces

    allResults := make([]ConftestFullScanResults, 0)

    repoData := NewRepoScrapperData()
    
    // Determine which repositories to scan
    if len(conftest.Repositories) == 0 {
        // Scan all repositories
        conftest.Repositories = make([]string, 0, len(repoData.Repos))
        for k := range repoData.Repos {
            conftest.Repositories = append(conftest.Repositories, k)
        }
    }

    count := 0
    for _, repoName := range conftest.Repositories {
        count++
        repo, found := repoData.Repos[repoName]
        if found == false {
            fmt.Printf("Unknown repository: %s\n", repoName)
            continue
        }

        
        fmt.Printf("(%d/%d) Processing Repo: %s\n", count, len(conftest.Repositories), repo.Name)
        targetRepo := fmt.Sprintf("%s/%s", conftest.BaseRepositoryLocation, repo.Name)

        files := conftest.getSelectedFiles(targetRepo)

        results := conftest.scanFiles(&files)

        record := ConftestFullScanResults {
            Repo: repo.Name,
            Team: repo.Team,
            Jira: repo.JiraProject,
            Status: repo.Status,
            IsLive: repo.IsLive,
            Results: results,
        }

        allResults = append(allResults, record)
    }

    return allResults
}

func (conftest *ConftestService) scanFiles (files *[]string) *[]ConftestResult {
    results := make([]ConftestResult, 0)

    count := 0
    for _, file := range *files {
        count++
        //fmt.Printf("(%d/%d) Scanning File: %s\n", count, len(*files), file)
        //command, args := conftest.getScanCommand(file)

        //fmt.Printf("Running Command: %s %v\n", command, args)

        cmd := conftest.getCommandObject(file)
        //cmd := exec.Command("ls", "-a")

        out, err := cmd.CombinedOutput()
        // if cmd.ProcessState.ExitCode() != 0 {
        //     fmt.Printf("Error Code `%d`: %s \n", cmd.ProcessState.ExitCode(), string(out))
        // }
        //fmt.Printf("Raw output: %s", string(out))
        //utils.CheckError(err, fmt.Sprintf("Running Conftest Command: %s %s", command, args))
        if err != nil {
            newRecord := ConftestResult {
                Level: "ERROR",
                Filepath: file,
                Message: string(out),
            }
            results = append(results, newRecord)
        } else {
            // TODO: Figure out why the --output "json" is not working
            // var scanResults ConftestScanResult
            // json.Unmarshal(out, &scanResults)
            conftest.processOutput(string(out), &results)
        }
    }

    return &results
}

// TODO: Was going to name it "GetCommand" but it looks like there might already be something with that name, so need to look it up
func (conftest *ConftestService) getCommandObject(file string) *exec.Cmd {
    // TODO: Figure out why "--output", " "json" doesn't work
    cmd := exec.Command("conftest", "test", "--no-color", "--no-fail", "--policy", conftest.PoliciesLocation)
    //, "--all-namespaces", file)
    if len(conftest.Namespaces) == 0 {
        cmd.Args = append(cmd.Args, "--all-namespaces")
    } else {
        for _, namespace := range conftest.Namespaces {
            cmd.Args = append(cmd.Args, "--namespace")
            cmd.Args = append(cmd.Args, namespace)
        }
    }
    cmd.Args = append(cmd.Args, file)

    //fmt.Printf("Final Command: %v\n", cmd)

    return cmd
}

type ConftestResult struct {
    Level string
    Filepath string
    Namespace string
    Message string
}

func (conftest *ConftestService) processOutput(output string, results *[]ConftestResult) {
    lines := strings.Split(output, "\n")

    for _, line := range lines {
        parts := strings.Split(line, " - ")
        if len(parts) == 1 {
            continue
        }



        if len(parts) != 4 {
            //panic(fmt.Sprintf("Output line is not made of 4 parts: %s", line))
            newRecrod := ConftestResult {
                Level: "UNKNOWN",
                Filepath: "UNKNOWN",
                Namespace: "UNKNOWN",
                Message: line,
            }
            *results = append(*results, newRecrod)
            continue
        }

        switch parts[0] {
        case "?": // Ignore
        case "WARN", "FAIL":
            newRecrod := ConftestResult {
                Level: parts[0],
                Filepath: parts[1],
                Namespace: parts[2],
                Message: parts[3],
            }
            *results = append(*results, newRecrod)
        }
    }
}

// GetSelectedFiles will identify all files that match the specified pattern in the target directory and all sub directories.
func (conftest *ConftestService) getSelectedFiles(targetBaseDirectory string) []string {
    selectedFiles := make([]string, 0)

    fileRegex, err := regexp.Compile(conftest.TargetFilePattern)
    utils.CheckError(err, "Conftest: Creating File Pattern regexp")

    ignoreRegex, err := regexp.Compile(conftest.IgnorePattern)
    utils.CheckError(err, "Conftest: Creating Ignore Pattern regexp")

    conftest.getSelectedFilesRecursive(targetBaseDirectory, *fileRegex, *ignoreRegex, &selectedFiles)

    return selectedFiles
}

// A recursive function that searches the current and all sub directories for files that match the target pattern
func (conftest *ConftestService) getSelectedFilesRecursive(dir string, fileRegex regexp.Regexp, ignoreRegex regexp.Regexp, selectedFiles *[]string) {
    files, err := ioutil.ReadDir(dir)
    utils.CheckError(err, "Conftest: Reading files from target directory")

    for _, file := range files {
        //fmt.Println(file.Name(), file.IsDir(), fileRegex.MatchString(file.Name()))

        if file.IsDir() == true {
            nextDir := fmt.Sprintf("%s/%s", dir, file.Name())
            conftest.getSelectedFilesRecursive(nextDir, fileRegex, ignoreRegex, selectedFiles)
        } else if fileRegex.MatchString(file.Name()) == true {
            filepath := fmt.Sprintf("%s/%s", dir, file.Name())
            if ignoreRegex.MatchString(filepath) == false {
                *selectedFiles = append(*selectedFiles, filepath)
            }
        }
    }
}