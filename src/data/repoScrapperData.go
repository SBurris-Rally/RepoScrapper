package data

import (
    "bufio"
	"fmt"
    "encoding/csv"
    "encoding/json"
    "errors"
    "io/ioutil"
	"os"
	"regexp"
	"strings"
    "strconv"
	"time"

	utils "Sburris/reposcrapper/src/utils"
)

const DefaultGitHubOwner string = "AudaxHealthInc"
const DefaultGitHubType string = "org" // user or org
const DefaultReposLocation string = "/Users/stephen.burris/.reposcrapper/repos"
const DefaultCacheLocation string = "/Users/stephen.burris/.reposcrapper/cache"
const DefaultHomeFolder string = "/Users/stephen.burris/.reposcrapper"

const GitHubAPIBase string = "https://api.github.com/"
const GitHubAPITokenEnv string = "GITHUB_TOKEN"

var GitHubAPIToken = "EMPTY"

type RepoScrapperData struct {
	Repos map[string]*Repository
	GitAPI GithubAPI
	Settings Settings
	LastUpdated time.Time
}

func NewRepoScrapperData() *RepoScrapperData {
	myData := &RepoScrapperData{
		GitAPI: GithubAPI {
			APIBase: GitHubAPIBase,
			TokenEnvKey: GitHubAPITokenEnv,
			Token: os.Getenv(GitHubAPITokenEnv),
		},
		Settings: Settings{
			FileSystemBaseFolder: DefaultHomeFolder,
			OutputFile: fmt.Sprintf("%s%s", DefaultHomeFolder, "/db.raw.json"),
		},
	}

	if _, err := os.Stat(myData.Settings.OutputFile); os.IsNotExist(err) {
		utils.Log("Cache not found, retrieving Repositories")
		
		//data.GetRepoMetadata()
		//data.GetRepoTrees()
		myData.Repos = make(map[string]*Repository)

		utils.Log("Finished Processing Repos")
		//data.SaveRepos(outputFilename, false)
		//data.CreateFileList()
		
	} else {
		utils.Log("Data Found - loading from cache")
		myData.LoadData(myData.Settings.OutputFile)
		utils.Log("Finished")
	}

	return myData
}

func (data *RepoScrapperData) LoadData(location string) {
	content, err := ioutil.ReadFile(location)
	utils.CheckError(err, fmt.Sprintf("Loading Data from '%s'", location))

	json.Unmarshal(content, &data.Repos)
}

func (data *RepoScrapperData) IdentifyReposBasedOnRepomanValue(values []string) []string {
	pattern := fmt.Sprintf("(?i).*(%s).*", strings.Join(values, ")|("))
	regexPattern, err := regexp.Compile(pattern)
	utils.CheckError(err, "Creating RegEx Pattern")

	matches := make([]string, 0)
	for _, repo := range data.Repos {
		isJiraProject := regexPattern.MatchString(repo.JiraProject)
		isTeam := regexPattern.MatchString(repo.Team)
		if isJiraProject || isTeam {
			repoInfo := fmt.Sprintf("%s, %s, %s", repo.Name, repo.Team, repo.JiraProject)
			matches = append(matches, repoInfo)
		}
	}

	return matches;
}

func (data *RepoScrapperData) DisplaySummary() {
	// emptyRepos := 0
	// emptyRepoNames := ""
	// truncatedRepos := 0
	// truncatedRepoNames := ""
	// usedRepos := 0
	isPublic := 0
	isPrivate := 0
	isFork := 0
	createdLastWeek := 0
	updatedLastWeek := 0
	//totalFileCount := 0

	lastWeek := time.Now().AddDate(0, 0, -7)

	for _, repo := range data.Repos {
		// fileCount := len(repo.Files)
		// if fileCount == 0 {
		// 	if repo.WasTreeTruncated {
		// 		truncatedRepos++
		// 		truncatedRepoNames += fmt.Sprintf("%s ", repo.Name)
		// 	} else {
		// 		emptyRepos++
		// 		emptyRepoNames += fmt.Sprintf("%s ", repo.Name)
		// 	}
			
		// } else {
		// 	usedRepos++
		// }
		// totalFileCount += totalFileCount

		if repo.IsPrivate {
			isPrivate++
		} else {
			isPublic++
		}

		if repo.IsFork {
			isFork++
		}

		if repo.CreatedAt.After(lastWeek) {
			createdLastWeek++
		}

		if repo.UpdatedAt.After(lastWeek) {
			updatedLastWeek++
		}
	}
	// Number of Repositories
	fmt.Printf("Repository Count: %d\n", len(data.Repos))
	//fmt.Printf("Used Repos: %d\n", usedRepos)
	//fmt.Printf("Truncated Repos: %d (%s)\n", truncatedRepos, truncatedRepoNames)
	//fmt.Printf("Empty Repos: %d (%s)\n", emptyRepos, emptyRepoNames)
	fmt.Printf("Private Repos: %d\n", isPrivate)
	fmt.Printf("Public Repos: %d\n", isPublic)
	fmt.Printf("Fork Repos: %d\n", isFork)
	fmt.Printf("Created Last Week: %d\n", createdLastWeek)
	fmt.Printf("Updated Last Week: %d\n", updatedLastWeek)

	// Number of empty repositories
	// Number of files
}

func (data *RepoScrapperData) CreateFileList() {
	// TODO: create a filelist

	// filepath := fmt.Sprintf("%s%s", data.settings.fileSystemBaseFolder, "/fullfilepaths.txt")
	// f, err := os.Create(filepath)
	// checkError(err, "Creating Full Filepaths file")
	// defer f.Close()
	// writer := bufio.NewWriter(f)	
	// for _, r := range data.repos {
	// 	for _, f := range r.Files {
	// 		if f.Type == "blob"{
	// 			writer.WriteString(fmt.Sprintf("%s/%s\n", r.FullName, f.Path))
	// 		}
	// 	}
	// }
	// writer.Flush()
}

func (data *RepoScrapperData) UpdateLocalRepos() {
	if _, err := os.Stat(DefaultReposLocation); os.IsNotExist(err) { 
		os.MkdirAll(DefaultReposLocation, 0744)
	}
	
	// Get github Repos
	allGitRepos, _ := data.GitAPI.GetAllRepos()
	utils.Log(fmt.Sprintf("Total Github Repos: %d\n", len(allGitRepos)))

	// Identify new Repos
	newRepos := getMissingRepos(allGitRepos, data.Repos)
	newRepoCount := len(newRepos)
	utils.Log(fmt.Sprintf("New Repo Count: %d", newRepoCount))
	count := 0
	for _, repo := range newRepos {
		count++
		utils.Log(fmt.Sprintf("(%d/%d) Cloning: %s", count, newRepoCount, repo.Name))
		repo.Update();
		data.Repos[repo.Name] = repo
		// TODO: any other things to set?
	}

	// Identify updated Repos
	updatedRepos := getUpdatedRepos(data.Repos, allGitRepos)
	updatedReposCount := len(updatedRepos)	
	utils.Log(fmt.Sprintf("Updated Repo Count: %d", updatedReposCount))
	count = 0
	for _, repo := range updatedRepos {
		count++
		utils.Log(fmt.Sprintf("(%d/%d) Updating: %s", count, updatedReposCount, repo.Name))
		repo.Update()
		
		// TODO: make proper function
		r := data.Repos[repo.Name]
		r.UpdatedAt = repo.UpdatedAt
	}

	// Identify deleted Repos
	deletedRepos := getMissingRepos(data.Repos, allGitRepos)
	deletedReposCount := len(deletedRepos)
	utils.Log(fmt.Sprintf("Deleted Repo Count: %d", deletedReposCount))
	count = 0
	for _, repo := range deletedRepos {
		count++
		utils.Log(fmt.Sprintf("(%d/%d) Deleting: %s", count, deletedReposCount, repo.Name))
		repo.Delete()
		delete(data.Repos, repo.Name)
	}

	data.LastUpdated = time.Now()
}

func (data *RepoScrapperData) IdentifyNoncompliantPullRequests() {
	pattern := "\\b[A-Za-z]{2,}-[0-9]{1,}\\b"
	regex, err := regexp.Compile(pattern)
	utils.CheckError(err, "Failed to complie PR regex")	

	prMetaData := make([]PrAnalysis, 0, 105000)
	for _, repo := range data.Repos {
		if repo.IsLive != "yes" {
			continue
		}

		prs, _ := data.GitAPI.GetPullRequests(repo.Name)
		for _, pullRequest := range prs {
			title := pullRequest.Title
			title = strings.ReplaceAll(title, "WIP", "")
			title = strings.ReplaceAll(title, "DNM", "")
			ticket := regex.FindAllString(title, 1)

			newRecord := PrAnalysis {
				RepoName: repo.Name,
				Status: repo.Status,
				Team: repo.Team,
				JiraProject: repo.JiraProject,
				IsLive: repo.IsLive,
				PrTitle: pullRequest.Title,
				CreatedAt: pullRequest.CreatedAt,
				UpdatedAt: pullRequest.UpdatedAt,
				ClosedAt: pullRequest.ClosedAt,
				MergedAt: pullRequest.MergedAt,
				TicketFound: len(ticket) > 0,
				Ticket: strings.Join(ticket, ","),
			}

			prMetaData = append(prMetaData, newRecord)
		}
	}
	
	//fmt.Println(prMetaData)

	// Export to CSV
	csvFilename := "PullRequests.csv"
	csvFile, err := os.Create(csvFilename)
	utils.CheckError(err, "creating Pull Request csv file.")
	defer csvFile.Close()

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()
	prMetaDataHeaders := []string {"RepoName", "Status", "Team", "JiraProject", "IsLive", "PrTitle", "CreatedAt", "UpdatedAt", "ClosedAt", "MergedAt", "TicketFound", "Ticket", }

	timeFormat := "2006-01-02 15:04:05"
	csvWriter.Write(prMetaDataHeaders)
	for _, pr := range prMetaData {
		csvRecord := []string {
			pr.RepoName,
			pr.Status,
			pr.Team,
			pr.JiraProject,
			pr.IsLive,
			pr.PrTitle,
			pr.CreatedAt.Format(timeFormat),
			pr.UpdatedAt.Format(timeFormat),
			pr.ClosedAt.Format(timeFormat),
			pr.MergedAt.Format(timeFormat),
			strconv.FormatBool(pr.TicketFound),
			pr.Ticket,
		}
		csvWriter.Write(csvRecord)
	}

	// Export to JSON
	jsonFilename := "PullRequests.json"
	jsonFile, err := os.Create(jsonFilename)
	utils.CheckError(err, "creating Pull Request json file.")
	defer jsonFile.Close()

	jsonData, err := json.MarshalIndent(prMetaData, "", "  ")
	utils.CheckError(err, "marshalling pr meta data")
	ioutil.WriteFile(jsonFilename, jsonData, 0600)


	// file, err := os.Create("invalidPrs.txt")
	// checkError(err, "Creating invalidPrs.txt")
	// defer file.Close()

	// for _, pr := range invalidPrs {
	// 	_, err := file.WriteString(fmt.Sprintf("%s\n", pr))
	// 	checkError(err, "Failed to write to file")
	// }
	
}

func (data *RepoScrapperData) UpdateMetadataFromRepoman() {
	utils.Log("Reading Repoman Files")

	repomanFiles := []string {"repoman.yml", "repoman.yaml", ".repoman.yml", ".repoman.yaml"}
	repomanIgnoreLines := []string {"---", "repository:", "testing:", "tags:", "dependencies:"}
	repomanSections := []string {"jira", "live", "status", "team", "slack", "slackroom", "jira_triaged", "jira-triaged", "jira_untriaged", "source", "project-id", "suite-id", "collaboratorsgroups", "name", "lead", "type", "tag", "product", "pillar", "- tag", "- name", "key", "owner"}
	//repomanLiveValues := []string{"yes", "\"yes\"", "'yes'", "no", "'no'", "never", "'never'", "\"never\""}
	for _, repo := range data.Repos {
		targetDir := fmt.Sprintf("%s/%s", DefaultReposLocation, repo.Name)
		for _, repomanFile := range repomanFiles {
			repomanFilePath := fmt.Sprintf("%s/%s", targetDir, repomanFile)
			if _, err := os.Stat(repomanFilePath); errors.Is(err, os.ErrNotExist) {
				continue
			}
			//Log(fmt.Sprintf("Found Repoman file in repo: %s", repo.Name))

			// Read File
			file, err := os.Open(repomanFilePath)
			utils.CheckError(err, "Reading Repoman file")

			scanner := bufio.NewScanner(file)

			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				skip := false
				for _, ignoreElement := range repomanIgnoreLines {
					if strings.Contains(line, ignoreElement) {
						skip = true
					}
				}

				if len(line) == 0 {
					skip = true
				}

				if skip {
					continue
				}

				parts := strings.Split(line, ":")
				if len(parts) == 1 {
					utils.Log(fmt.Sprintf("Subcategory not found: '%s'", line))
					continue
				}

				section := strings.ToLower(strings.TrimSpace(parts[0]))
				knownElement := utils.Contains(repomanSections, section)
				if !knownElement {
					utils.Log(fmt.Sprintf("Unknown Section: '%s'", line))
				}

				// Status string
				if section == "status" {
					value := strings.TrimSpace(parts[1])
					v := repomanStripQuotes(value)
					repo.Status = v
				}
				
				if section == "team" || section == "owner" {
					value := strings.TrimSpace(parts[1])
					v := repomanStripQuotes(value)
					repo.Team = v
				}

				if section == "jira" || section == "key" {
					value := strings.TrimSpace(parts[1])
					v := repomanStripQuotes(value)
					repo.JiraProject = v
				}

				if section == "live" {
					value := strings.ToLower(strings.TrimSpace(parts[1]))
					v := repomanStripQuotes(value)
					repo.IsLive = v
				}
			}
		}
	}
}

func (data *RepoScrapperData) GetRepoMetadata() {
	count := 0
	var repos []*Repository
	var err error

	//allGitRepos, _ := data.gitapi.GetAllRepos()

	// TODO: Check for new, need to be updated, and deleted repos.
	for {
		count++
		repos, err = data.GitAPI.GetRepos(count)
		
		// TODO: Why is this here... what was the reason?
		for err != nil {
			repos, err = data.GitAPI.GetRepos(count)
		}

		if repos != nil {
			// TODO: this is where the update action needs to be determined
			for _, repo := range repos {
				data.Repos[repo.Name] = repo
			}
		} else {
			break // no records, implies we are at the end of the repos
		}
	}
}

func (data *RepoScrapperData) SaveRepos(isFormatted bool) {
	utils.Log("Saving Info to JSON File")

	var file []byte
	var err error
	if isFormatted {
		file, err = json.MarshalIndent(data.Repos, "", "  ")
	} else {
		file, err = json.Marshal(data.Repos)
	}	
	utils.CheckError(err, "Marshalling Repos to JSON")

	utils.SaveFile(file, data.Settings.OutputFile)
}

func (data *RepoScrapperData) MapSnykProjects(inputFilePath string, outputFilePath string) {
	inputFiles := []string{"snyk_container_spring-webmvc_spring-webflux.csv", "snyk_temp_spring-webmvc_spring-webflux.csv"}
	
	print(inputFilePath)

	outputFile, err := os.Create(outputFilePath)
	utils.CheckError(err, fmt.Sprintf("Creating output file: %s", outputFilePath))
	defer outputFile.Sync()
	defer outputFile.Close()
	outputFile.WriteString("Repository, Dep Name, Dep Version, Location, Status, Team, Jira, Notes\n")

	for _, inputFilename := range inputFiles {
		inputFile, err := os.Open(inputFilename)
		utils.CheckError(err, fmt.Sprintf("Opening Snyk CSV for mapping purposes: %s", inputFilePath))
		defer inputFile.Close()

		reader := csv.NewReader(inputFile)
		payload, err := reader.ReadAll()
		utils.CheckError(err, "Reading Snyk CSV file")

		for index, row := range payload {
			if index == 0 {
				
				continue // Skip headers
			}

			projects := strings.Split(row[10], ",")
			for _, project := range projects {
				projectName := strings.Split(project, ":")[0]
				projectName = strings.TrimSpace(projectName)
				projectName = strings.Replace(projectName, "AudaxHealthInc/", "", 1)
				
				repo, isFound := data.Repos[projectName]
				
				if isFound {
					line := fmt.Sprintf("%s, %s, %s, %s, %s, %s, %s,\n", repo.Name, row[1], row[2], strings.TrimSpace(project), repo.Status, repo.Team, repo.JiraProject)
					outputFile.WriteString(line)
				} else {
					line := fmt.Sprintf(",%s,%s,%s,,,,\n", row[1], row[2], strings.TrimSpace(project))
					outputFile.WriteString(line)
				}
			}
		}
	}

	fmt.Println("Done")
}

// Check to see what repos are in the first set but not in the second
func getMissingRepos(firstSet map[string]*Repository, second map[string]*Repository) []*Repository {
	missingRepos := make([]*Repository, 0)

	for key, repo := range firstSet {
		_, found := second[key]
		if !found {
			missingRepos = append(missingRepos, repo)
		}
	}

	return missingRepos
}

func getUpdatedRepos(knownRepos map[string]*Repository, newRepos map[string]*Repository) []*Repository {
	updatedRepos := make([]*Repository, 0)

	for key, repo := range knownRepos {
		newRepo, found := newRepos[key]
		if found {
			if repo.UpdatedAt != newRepo.UpdatedAt {
				updatedRepos = append(updatedRepos, newRepo)
			}
		}
	}

	return updatedRepos
}

func repomanStripQuotes(value string) string {
	value = strings.ReplaceAll(value, "'", "")
	value = strings.ReplaceAll(value, "\"", "")

	return value
}
