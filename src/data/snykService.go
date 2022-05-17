package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	utils "Sburris/reposcrapper/src/utils"
)

const SnykAPIBase string = "https://snyk.io/api/v1"
const SnykTokenEnvKey = "SNYK_TOKEN"
const ScaOrgID string = "9bf47dc4-8c9d-4ee5-af4d-4270a19dd1c6"

type SnykService struct {
    APIBase string
    Token string
}

func CreateSnykService(apiBase string, tokenEnvKey string) *SnykService {
    snykService := SnykService {
        APIBase: apiBase,
        Token: os.Getenv(tokenEnvKey),
    }
    return &snykService
}

func (service *SnykService) GetRepos() []string {
    rawProjects := service.getProjectsForOrg(ScaOrgID)
    repositories := extractReposFromProjectslist(rawProjects)

    utils.SaveSliceOfStringsToFile("snykRepos.txt", repositories)

    return repositories
}

func (service *SnykService) getProjectsForOrg(orgID string) SnykProjectsListPayload {
    targetURL := fmt.Sprintf("%s/org/%s/projects", SnykAPIBase, orgID)

    client := &http.Client{}
    request, err := http.NewRequest("POST", targetURL, nil)
    utils.CheckError(err, "New Request for Snyk GET projects list")

    request.Header.Add("Accept", "application/json")
    request.Header.Add("Authorization", fmt.Sprintf("token %s", service.Token))

    response, err := client.Do(request)
    utils.CheckError(err, "Getting response from Snyk's GET projects list endpoint")

    defer response.Body.Close()
    bodyBytes, err := ioutil.ReadAll(response.Body)
    utils.CheckError(err, "Reading Snyk's GET project response")

    if response.StatusCode != 200 {
        utils.Log(fmt.Sprintf("Status Code is '%d' and body is \n%s", response.StatusCode, string(bodyBytes)))
    }

    var projects SnykProjectsListPayload
    json.Unmarshal(bodyBytes, &projects)

    return projects
}

func extractReposFromProjectslist(projects SnykProjectsListPayload) []string {
    repositories := make([]string, 0, len(projects.Projects))

    fileTypes := make([]string, 0)

    interestedInFiles := []string {"1.1.1", "test", "lib", "unlimited", "limited", "ext", "test", "newrelic"}

    for _, project := range projects.Projects {
        filenameStart := strings.Index(project.Name, ":")
        repository := project.Name
        if filenameStart >= 0 {
            repository = project.Name[:filenameStart]

            
            fileType := project.Name[filenameStart+1:]
            f := filepath.Base(fileType)
            if(!utils.Contains(fileTypes, f)) {
                fileTypes = append(fileTypes, f)
            }

            if utils.Contains(interestedInFiles, f) {
                fmt.Printf("Interested: %s\n", project.Name)
            }
        } else {
            fmt.Printf("No File: %s\n", project.Name)
        }
        
        if !utils.Contains(repositories, repository) {
            repositories = append(repositories, repository)
        }
    }

    utils.SaveSliceOfStringsToFile("fileTypes.txt", fileTypes)

    return repositories
}

/// ***********************************************************************************************
///       Data Structures
/// ***********************************************************************************************

type SnykProjectsListPayload struct {
    Org struct {
        Name string `json:"name"`
        ID   string `json:"id"`
    } `json:"org"`
    Projects []struct {
        Name                  string    `json:"name"`
        ID                    string    `json:"id"`
        Created               time.Time `json:"created"`
        Origin                string    `json:"origin"`
        Type                  string    `json:"type"`
        ReadOnly              bool      `json:"readOnly"`
        TestFrequency         string    `json:"testFrequency"`
        TotalDependencies     int       `json:"totalDependencies"`
        IssueCountsBySeverity struct {
            Low      int `json:"low"`
            Medium   int `json:"medium"`
            High     int `json:"high"`
            Critical int `json:"critical"`
        } `json:"issueCountsBySeverity"`
        RemoteRepoURL  string    `json:"remoteRepoUrl,omitempty"`
        LastTestedDate time.Time `json:"lastTestedDate"`
        ImportingUser  struct {
            ID       string `json:"id"`
            Name     string `json:"name"`
            Username string `json:"username"`
            Email    string `json:"email"`
        } `json:"importingUser"`
        IsMonitored bool `json:"isMonitored"`
        Owner       struct {
            ID       string `json:"id"`
            Name     string `json:"name"`
            Username string `json:"username"`
            Email    string `json:"email"`
        } `json:"owner"`
        Branch string `json:"branch"`
        Tags   []struct {
            Key   string `json:"key"`
            Value string `json:"value"`
        } `json:"tags"`
        ImageID  string `json:"imageId,omitempty"`
        ImageTag string `json:"imageTag,omitempty"`
    } `json:"projects"`
}