package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	//"log" 			 TODO: Look ingo
	"encoding/json"
	"io/ioutil"
	"net/http"

	"path/filepath"
	"regexp"
	"strconv"
)

// Settings to be moved into settings file
// settings file location = "~/.reposcrapper/settings.json"
const DEFAULT_GITHUB_OWNER string = "AudaxHealthInc"
const DEFAULT_GITHUB_TYPE string = "org" // user or org
const DEFAULT_REPOS_LOCATION string = "/Users/stephen.burris/.reposcrapper/repos"
const DEFAULT_CACHE_LOCATION string = "/Users/stephen.burris/.reposcrapper/cache"
const DEFAULT_HOME_FOLDER string = "/Users/stephen.burris/.reposcrapper"

const DISPLAY_FORMAT_TIME_ONLY_HIGH_PERCISION string = "15:04:05.000"
// const DISPLAY_FORMAT_TIME_ONLY string = "15:04:05"
// const DISPLAY_FORMAT_DATE_ONLY string = ""
// const DISPLAY_FORMAT_DATE_AND_TIME_SORTABLE string = "2021-08-31 15:04:05"
// const DISPLAY_FORMAT_DATE_AND_TIME_HUMAN_READABLE string = "31 Aug 2021 15:04:05"

const GITHUB_API_BASE string = "https://api.github.com/"
const GITHUB_API_TOKEN_ENV string = "GITHUB_TOKEN"

var GITHUB_API_TOKEN = "EMPTY"

func main() {
	outputFilename := fmt.Sprintf("%s%s", DEFAULT_HOME_FOLDER, "/db.raw.json")

	data := &Data{
		gitapi: githubApi {
			apiBase: GITHUB_API_BASE,
			tokenEnvKey: GITHUB_API_TOKEN_ENV,
			token: os.Getenv(GITHUB_API_TOKEN_ENV),
		},
		settings: Settings{
			fileSystemBaseFolder: DEFAULT_HOME_FOLDER,
		},
	}
	
	//data.GetRepoMetadata()

	if _, err := os.Stat(outputFilename); os.IsNotExist(err) {
		Log("Cache not found, retrieving Repositories")
		
		data.GetRepoMetadata()
		//data.GetRepoTrees()

		Log("Finished Processing Repos")
		data.SaveRepos(outputFilename, false)
		//data.CreateFileList()
		
	} else {
		Log("Data Found - loading from cache")
		data.LoadData(outputFilename)
		Log("Finished")
	}

	data.CloneAllRepos()

	//data.CreateFileList()
	
	//data.DisplaySummary()

	//temp := fmt.Sprintf("%s/%s", DEFAULT_HOME_FOLDER, "db.raw.json")
	//data.SaveRepos(temp, false)


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
}

func (data *Data) DownloadFiles(pattern string) {
	// Search all repos
	regexPattern, err := regexp.Compile(pattern)
	checkError(err, "Creating RegEx Pattern")

	found := 0
	//repoCount := len(repos)
	for _, repo := range data.repos {
		//Log(fmt.Sprintf("[%d/%d] Search Repo: %s", rIndex, repoCount, repo.FullName))
		for _, leaf := range repo.Files {
			if(leaf.Type == "blob") {
				filename := filepath.Base(leaf.Path)
				if regexPattern.MatchString(filename) {
					found++
					Log(fmt.Sprintf("Found Match: %s/%s", repo.FullName, leaf.Path))
					//continue
					fullFilename := filepath.Join(DEFAULT_CACHE_LOCATION, repo.FullName, leaf.Path)

					// Get file
					fileBytes, err := data.gitapi.GetFile(repo, *leaf)
					for fileBytes == nil && err == nil {
						fileBytes, err = data.gitapi.GetFile(repo, *leaf)
					}
	
					SaveFile(fileBytes, fullFilename)
				}
			}
		}
	}
	Log(fmt.Sprintf("Found: %d", found))
}

func (data *Data) DisplaySummary() {
	emptyRepos := 0
	emptyRepoNames := ""
	truncatedRepos := 0
	truncatedRepoNames := ""
	usedRepos := 0
	isPublic := 0
	isPrivate := 0
	isFork := 0
	createdLastWeek := 0
	updatedLastWeek := 0
	totalFileCount := 0

	lastWeek := time.Now().AddDate(0, 0, -7)

	for _, repo := range data.repos {
		fileCount := len(repo.Files)
		if fileCount == 0 {
			if repo.WasTreeTruncated {
				truncatedRepos++
				truncatedRepoNames += fmt.Sprintf("%s ", repo.Name)
			} else {
				emptyRepos++
				emptyRepoNames += fmt.Sprintf("%s ", repo.Name)
			}
			
		} else {
			usedRepos++
		}
		totalFileCount += totalFileCount

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
	fmt.Printf("Repository Count: %d\n", len(data.repos))
	fmt.Printf("Used Repos: %d\n", usedRepos)
	fmt.Printf("Truncated Repos: %d (%s)\n", truncatedRepos, truncatedRepoNames)
	fmt.Printf("Empty Repos: %d (%s)\n", emptyRepos, emptyRepoNames)
	fmt.Printf("Private Repos: %d\n", isPrivate)
	fmt.Printf("Public Repos: %d\n", isPublic)
	fmt.Printf("Fork Repos: %d\n", isFork)
	fmt.Printf("Created Last Week: %d\n", createdLastWeek)
	fmt.Printf("Updated Last Week: %d\n", updatedLastWeek)

	// Number of empty repositories
	// Number of files
}

func (data *Data) CreateFileList() {
	filepath := fmt.Sprintf("%s%s", data.settings.fileSystemBaseFolder, "/fullfilepaths.txt")
	f, err := os.Create(filepath)
	checkError(err, "Creating Full Filepaths file")
	defer f.Close()
	writer := bufio.NewWriter(f)	
	for _, r := range data.repos {
		for _, f := range r.Files {
			if f.Type == "blob"{
				writer.WriteString(fmt.Sprintf("%s/%s\n", r.FullName, f.Path))
			}
		}
	}
	writer.Flush()
}

func (data *Data) GetRepoMetadata() {
	count := 0
	var repos []Repository
	var err error
	for {
		count++
		repos, err = data.gitapi.GetRepos(count)
		for err != nil {
			repos, err = data.gitapi.GetRepos(count)
		}

		if repos != nil {
			data.repos = append(data.repos, repos...)
		} else {
			break // no records, implies we are at the end of the repos
		}
	}
}

func (git githubApi) GetRepos(page int) (out []Repository, err error) {
	Log(fmt.Sprintf("Performing Http Get - Page %d", page))

	client := &http.Client{}
	// https://api.github.com/search/repositories?q=user:sburris fork:true&per_page=100   // TODO: why is this rate-limit at 60?
	// https://api.github.com/users/Sburris/repos?per_page=100
	// req, err := http.NewRequest("GET", "https://api.github.com/search/repositories?q=user:sburris%20fork:true&per_page=100", nil)
	// TODO: Currently only getting one page, need to refactor to grab all
	url := fmt.Sprintf("https://api.github.com/orgs/AudaxHealthInc/repos?per_page=100&page=%d", page)
	req, err := http.NewRequest("GET", url, nil)
	checkError(err, "New Request for GET user repo")

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", git.token))

	//Log("Performing Call")
	resp, err := client.Do(req)
	checkError(err, "client.Do")
	
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	checkError(err, "Ready repo response body")
	
	if resp.StatusCode != 200 {
		PrintResponseHeaders(*resp)
		SleepUntil(resp.Header["X-Ratelimit-Reset"][0])
		return nil, errors.New("Timeout")
	}

	var repoResponse GithubRepoListGet
	json.Unmarshal(bodyBytes, &repoResponse)

	for _, repo := range repoResponse {
		newRepo := Repository {
			Id: repo.ID,
			NodeId: repo.NodeID,
			Name: repo.Name,
			FullName: repo.FullName,
			IsPrivate: repo.Private,
			IsFork: repo.Fork,
			//IsArchived: repo.Archived,
			HtmlUrl: repo.HTMLURL,
			Description: repo.Description,
			CreatedAt: repo.CreatedAt,
			UpdatedAt: repo.UpdatedAt,
			PushedAt: repo.PushedAt,
			DefaultBranch: repo.DefaultBranch,
			SshUrl: repo.SSHURL,
		}
		//newRepo.Files = make([]RepositoryFile, 0, 200)
		out = append(out, newRepo)
	}

	return out, nil
}

func PrintResponseHeaders(resp http.Response) {
	fmt.Printf("Response Status Code: %d\n", resp.StatusCode)
	fmt.Printf("X-RateLimit-Limit: %s\n", resp.Header["X-Ratelimit-Limit"][0])
	fmt.Printf("X-RateLimit-Remaining: %s\n", resp.Header["X-Ratelimit-Remaining"][0])
	fmt.Printf("X-RateLimit-Used: %s\n", resp.Header["X-Ratelimit-Used"][0])
	fmt.Printf("X-RateLimit-Reset: %s\n", resp.Header["X-Ratelimit-Reset"][0])
}

func (data *Data) GetRepoTrees() {
	
	repoTotal := len(data.repos)
	for i := range data.repos {
		repo := &data.repos[i]
		treeResponse := data.gitapi.GetGithubTree(repo.Name, repo.DefaultBranch, true, i, repoTotal)
		for treeResponse == nil {
			treeResponse = data.gitapi.GetGithubTree(repo.Name, repo.DefaultBranch, true, i, repoTotal)
		}

		if treeResponse.Truncated {
			Log(fmt.Sprintf("TREE Truncated: %s", repo.Name))
			repo.WasTreeTruncated = true
			continue // don't save the files, currently.  The two projects that do this are new and look to be test projects, plus I don't want the extra 30MB need to house all of their dependencies
			// TODO: possibly put in exclusion folders to not add stuff to our records, this npm and other folders would be good for that.
		}
		repo.WasTreeTruncated = false

		//repo.Files = make([]RepositoryFile, 0, len(treeResponse.Tree))
		for _, leaf := range treeResponse.Tree {

			newLeaf := RepositoryFile {
				Path: leaf.Path,
				Mode: leaf.Mode,
				Type: leaf.Type,
				Sha: leaf.Sha,
				Size: leaf.Size,
				Url: leaf.URL,
			}
			//Log(fmt.Sprintf("New Leave Added: [%s] %s", repo.Name, newLeaf.Path))
			repo.Files = append(repo.Files, &newLeaf)
			//repos[i].Files = append(repos[i].Files, newLeaf)
		}
	}
}

func (github *githubApi) GetGithubTree(repoName string, defaultBranch string, isRecursive bool, index int, total int) *GithubTreeResponse {
	client := &http.Client{}
	treeRequestUrl := fmt.Sprintf("https://api.github.com/repos/AudaxHealthInc/%s/git/trees/%s", repoName, defaultBranch)
	if isRecursive {
		treeRequestUrl = treeRequestUrl + "?recursive=1"
	}

	//treeRequestUrl := "https://api.github.com/repos/SBurris/SecurityNotes/git/trees/master?recursive=1"
	Log(fmt.Sprintf("[%d/%d]Retrieving Tree for Repo: %s (%s)",index, total, repoName, treeRequestUrl))
	treeReq, treeErr := http.NewRequest("GET", treeRequestUrl, nil)
	checkError(treeErr, "New Request for GET user repo")

	treeReq.Header.Add("Accept", "application/json")
	treeReq.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", github.token))

	treeResp, err := client.Do(treeReq)
	checkError(err, "client.Do")
	fmt.Printf("X-RateLimit-Remaining: %s\n", treeResp.Header["X-Ratelimit-Remaining"][0])

	defer treeResp.Body.Close()
	treeBodyBytes, err := ioutil.ReadAll(treeResp.Body)
	checkError(err, "Ready repo response body")

	if treeResp.StatusCode != 200 && treeResp.StatusCode != 404 && treeResp.StatusCode != 409 {
		PrintResponseHeaders(*treeResp)
		SleepUntil(treeResp.Header["X-Ratelimit-Reset"][0])
		return nil
	}

	var treeResponse GithubTreeResponse
	json.Unmarshal(treeBodyBytes, &treeResponse)

	return &treeResponse
}


func (data *Data) SaveRepos(path string, isFormatted bool) {
	Log("Saving Info to JSON File")

	var file []byte
	var err error
	if isFormatted {
		file, err = json.MarshalIndent(data.repos, "", "  ")
	} else {
		file, err = json.Marshal(data.repos)
	}	
	checkError(err, "Marshalling Repos to JSON")

	SaveFile(file, path)
}

func (data *Data) CloneAllRepos() {
	Log("Cloning All Repos")

	// command: git clone <repo> <location>

	repoCount := len(data.repos)
	for i, repo := range data.repos {
		targetDir := fmt.Sprintf("%s/%s", DEFAULT_REPOS_LOCATION, repo.Name)

		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			fmt.Printf("Cloning Repo %s (%d/%d)\n", repo.Name, i+1, repoCount)

			cmd := exec.Command("git", "clone", repo.SshUrl, targetDir)
			//cmd.Path = DEFAULT_REPOS_LOCATION
			//fmt.Print(cmd)
			
			err := cmd.Run()
			checkError(err, "Failed cloning repo")
		} else {
			fmt.Printf("Pulling Repo %s (%d/%d)\n", repo.Name, i+1, repoCount)

			cmd := exec.Command("git", "-C", targetDir, "pull")
			//fmt.Print(cmd)

			err := cmd.Run()
			checkError(err, "Failed to pull repo")
		}
	}
}

func (data *Data) LoadData(location string) {
	content, err := ioutil.ReadFile(location)
	checkError(err, fmt.Sprintf("Loading Data from '%s'", location))

	json.Unmarshal(content, &data.repos)
}

func (github *githubApi) GetFile(repo Repository, file RepositoryFile) ([]byte, error) {
	fileUrl := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", DEFAULT_GITHUB_OWNER, repo.Name, file.Path)

	client := &http.Client{}
	fileReq, err := http.NewRequest("GET", fileUrl, nil)
	checkError(err, "New Request for GET user repo")

	fileReq.Header.Add("Accept", "application/vnd.github.VERSION.raw")
	fileReq.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", github.token))

	fileResp, err := client.Do(fileReq)
	checkError(err, "client.Do")
	
	defer fileResp.Body.Close()
	fileBytes, err := ioutil.ReadAll(fileResp.Body)
	checkError(err, "Ready repo response body")

	if fileResp.StatusCode == 404 {
		PrintResponseHeaders(*fileResp)
		return nil, errors.New("404")
	}

	if fileResp.StatusCode == 403 {
		PrintResponseHeaders(*fileResp)
		SleepUntil(fileResp.Header["X-Ratelimit-Reset"][0])
		return nil, nil
	}

	if fileResp.StatusCode != 200 {
		PrintResponseHeaders(*fileResp)
		return nil, errors.New(fmt.Sprintf("%d", fileResp.StatusCode))
	}

	return fileBytes, nil
}

func SleepUntil(reset string) {
	expireTime, _ := strconv.Atoi(reset)
	resetTime := time.Unix(int64(expireTime), 0).Add(time.Second * 30)
	waitPeriod := resetTime.Unix() - time.Now().Unix()

	dur := time.Until(resetTime)
	Log(fmt.Sprintf("Sleeping For %s [%d]", dur.String(), waitPeriod))
	time.Sleep(time.Until(resetTime))
}

// func createRepoMap(in []Repository) (out map[string]*Repository) {
// 	for i:= 0; i < len(in); i++ {
// 		out[in[i].FullName] = &in[i]
// 	}
// 	return
// }

func Log(msg string) {
	fmt.Printf("[%s] %s\n", time.Now().Format(DISPLAY_FORMAT_TIME_ONLY_HIGH_PERCISION), msg)
}

func checkError(err error, msg string) {
	if err != nil {
		Log(fmt.Sprintf("[ERROR] %s\n%v", msg, err))
		os.Exit(1)
	}
}

func SaveFile(bytes []byte, filename string) {
	targetDir := filepath.Dir(filename)

	if _, err := os.Stat(targetDir); os.IsNotExist(err) { 
		os.MkdirAll(targetDir, 0744) // Create your file
	}
	
	Log(fmt.Sprintf("Saving file to: %s", filename))
	errWrite := ioutil.WriteFile(filename, bytes, 0644)
	checkError(errWrite, "Error writing file")
}

type Data struct {
	repos []Repository
	gitapi githubApi
	settings Settings
	lastUpdated time.Time
}

type Settings struct {
	fileSystemBaseFolder string
}

type githubApi struct {
	apiBase string
	tokenEnvKey string
	token string
}

type Repository struct {
	Id int
	NodeId string
	Name string
	FullName string
	IsPrivate bool
	IsFork	bool
	HtmlUrl string
	Description string
	CreatedAt time.Time
	UpdatedAt time.Time
	PushedAt time.Time
	DefaultBranch string
	Files []*RepositoryFile
	NeedToProcess bool
	WasTreeTruncated bool
	SshUrl string
}

type RepositoryFile struct {
	Path string
	Mode string
	Type string
	Sha string
	Size int
	Url string
	DownloadedDate time.Time
}

// https://mholt.github.io/json-to-go/
type GithubSearchResponse struct {
	TotalCount        int  `json:"total_count"`
	IncompleteResults bool `json:"incomplete_results"`
	Items             []struct {
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Private  bool   `json:"private"`
		Owner    struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"owner"`
		HTMLURL          string      `json:"html_url"`
		Description      string      `json:"description"`
		Fork             bool        `json:"fork"`
		URL              string      `json:"url"`
		ForksURL         string      `json:"forks_url"`
		KeysURL          string      `json:"keys_url"`
		CollaboratorsURL string      `json:"collaborators_url"`
		TeamsURL         string      `json:"teams_url"`
		HooksURL         string      `json:"hooks_url"`
		IssueEventsURL   string      `json:"issue_events_url"`
		EventsURL        string      `json:"events_url"`
		AssigneesURL     string      `json:"assignees_url"`
		BranchesURL      string      `json:"branches_url"`
		TagsURL          string      `json:"tags_url"`
		BlobsURL         string      `json:"blobs_url"`
		GitTagsURL       string      `json:"git_tags_url"`
		GitRefsURL       string      `json:"git_refs_url"`
		TreesURL         string      `json:"trees_url"`
		StatusesURL      string      `json:"statuses_url"`
		LanguagesURL     string      `json:"languages_url"`
		StargazersURL    string      `json:"stargazers_url"`
		ContributorsURL  string      `json:"contributors_url"`
		SubscribersURL   string      `json:"subscribers_url"`
		SubscriptionURL  string      `json:"subscription_url"`
		CommitsURL       string      `json:"commits_url"`
		GitCommitsURL    string      `json:"git_commits_url"`
		CommentsURL      string      `json:"comments_url"`
		IssueCommentURL  string      `json:"issue_comment_url"`
		ContentsURL      string      `json:"contents_url"`
		CompareURL       string      `json:"compare_url"`
		MergesURL        string      `json:"merges_url"`
		ArchiveURL       string      `json:"archive_url"`
		DownloadsURL     string      `json:"downloads_url"`
		IssuesURL        string      `json:"issues_url"`
		PullsURL         string      `json:"pulls_url"`
		MilestonesURL    string      `json:"milestones_url"`
		NotificationsURL string      `json:"notifications_url"`
		LabelsURL        string      `json:"labels_url"`
		ReleasesURL      string      `json:"releases_url"`
		DeploymentsURL   string      `json:"deployments_url"`
		CreatedAt        time.Time   `json:"created_at"`
		UpdatedAt        time.Time   `json:"updated_at"`
		PushedAt         time.Time   `json:"pushed_at"`
		GitURL           string      `json:"git_url"`
		SSHURL           string      `json:"ssh_url"`
		CloneURL         string      `json:"clone_url"`
		SvnURL           string      `json:"svn_url"`
		Homepage         interface{} `json:"homepage"`
		Size             int         `json:"size"`
		StargazersCount  int         `json:"stargazers_count"`
		WatchersCount    int         `json:"watchers_count"`
		Language         string      `json:"language"`
		HasIssues        bool        `json:"has_issues"`
		HasProjects      bool        `json:"has_projects"`
		HasDownloads     bool        `json:"has_downloads"`
		HasWiki          bool        `json:"has_wiki"`
		HasPages         bool        `json:"has_pages"`
		ForksCount       int         `json:"forks_count"`
		MirrorURL        interface{} `json:"mirror_url"`
		Archived         bool        `json:"archived"`
		Disabled         bool        `json:"disabled"`
		OpenIssuesCount  int         `json:"open_issues_count"`
		License          interface{} `json:"license"`
		Forks            int         `json:"forks"`
		OpenIssues       int         `json:"open_issues"`
		Watchers         int         `json:"watchers"`
		DefaultBranch    string      `json:"default_branch"`
		Permissions      struct {
			Admin bool `json:"admin"`
			Push  bool `json:"push"`
			Pull  bool `json:"pull"`
		} `json:"permissions"`
		Score float64 `json:"score"`
	} `json:"items"`
}

type GithubUserRepoResponse []struct {
	ID       int    `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"owner"`
	HTMLURL          string      `json:"html_url"`
	Description      interface{} `json:"description"`
	Fork             bool        `json:"fork"`
	URL              string      `json:"url"`
	ForksURL         string      `json:"forks_url"`
	KeysURL          string      `json:"keys_url"`
	CollaboratorsURL string      `json:"collaborators_url"`
	TeamsURL         string      `json:"teams_url"`
	HooksURL         string      `json:"hooks_url"`
	IssueEventsURL   string      `json:"issue_events_url"`
	EventsURL        string      `json:"events_url"`
	AssigneesURL     string      `json:"assignees_url"`
	BranchesURL      string      `json:"branches_url"`
	TagsURL          string      `json:"tags_url"`
	BlobsURL         string      `json:"blobs_url"`
	GitTagsURL       string      `json:"git_tags_url"`
	GitRefsURL       string      `json:"git_refs_url"`
	TreesURL         string      `json:"trees_url"`
	StatusesURL      string      `json:"statuses_url"`
	LanguagesURL     string      `json:"languages_url"`
	StargazersURL    string      `json:"stargazers_url"`
	ContributorsURL  string      `json:"contributors_url"`
	SubscribersURL   string      `json:"subscribers_url"`
	SubscriptionURL  string      `json:"subscription_url"`
	CommitsURL       string      `json:"commits_url"`
	GitCommitsURL    string      `json:"git_commits_url"`
	CommentsURL      string      `json:"comments_url"`
	IssueCommentURL  string      `json:"issue_comment_url"`
	ContentsURL      string      `json:"contents_url"`
	CompareURL       string      `json:"compare_url"`
	MergesURL        string      `json:"merges_url"`
	ArchiveURL       string      `json:"archive_url"`
	DownloadsURL     string      `json:"downloads_url"`
	IssuesURL        string      `json:"issues_url"`
	PullsURL         string      `json:"pulls_url"`
	MilestonesURL    string      `json:"milestones_url"`
	NotificationsURL string      `json:"notifications_url"`
	LabelsURL        string      `json:"labels_url"`
	ReleasesURL      string      `json:"releases_url"`
	DeploymentsURL   string      `json:"deployments_url"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	PushedAt         time.Time   `json:"pushed_at"`
	GitURL           string      `json:"git_url"`
	SSHURL           string      `json:"ssh_url"`
	CloneURL         string      `json:"clone_url"`
	SvnURL           string      `json:"svn_url"`
	Homepage         interface{} `json:"homepage"`
	Size             int         `json:"size"`
	StargazersCount  int         `json:"stargazers_count"`
	WatchersCount    int         `json:"watchers_count"`
	Language         interface{} `json:"language"`
	HasIssues        bool        `json:"has_issues"`
	HasProjects      bool        `json:"has_projects"`
	HasDownloads     bool        `json:"has_downloads"`
	HasWiki          bool        `json:"has_wiki"`
	HasPages         bool        `json:"has_pages"`
	ForksCount       int         `json:"forks_count"`
	MirrorURL        interface{} `json:"mirror_url"`
	Archived         bool        `json:"archived"`
	Disabled         bool        `json:"disabled"`
	OpenIssuesCount  int         `json:"open_issues_count"`
	License          struct {
		Key    string `json:"key"`
		Name   string `json:"name"`
		SpdxID string `json:"spdx_id"`
		URL    string `json:"url"`
		NodeID string `json:"node_id"`
	} `json:"license"`
	Forks         int    `json:"forks"`
	OpenIssues    int    `json:"open_issues"`
	Watchers      int    `json:"watchers"`
	DefaultBranch string `json:"default_branch"`
	Permissions   struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	} `json:"permissions"`
}

type GithubTreeResponse struct {
	Sha  string `json:"sha"`
	URL  string `json:"url"`
	Tree []struct {
		Path string `json:"path"`
		Mode string `json:"mode"`
		Type string `json:"type"`
		Sha  string `json:"sha"`
		Size int    `json:"size,omitempty"`
		URL  string `json:"url"`
	} `json:"tree"`
	Truncated bool `json:"truncated"`
}

// API Response: /orgs/{org}/repos
type GithubRepoListGet []struct {
	ID       int    `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"owner"`
	Private          bool        `json:"private"`
	HTMLURL          string      `json:"html_url"`
	Description      string      `json:"description"`
	Fork             bool        `json:"fork"`
	URL              string      `json:"url"`
	ArchiveURL       string      `json:"archive_url"`
	AssigneesURL     string      `json:"assignees_url"`
	BlobsURL         string      `json:"blobs_url"`
	BranchesURL      string      `json:"branches_url"`
	CollaboratorsURL string      `json:"collaborators_url"`
	CommentsURL      string      `json:"comments_url"`
	CommitsURL       string      `json:"commits_url"`
	CompareURL       string      `json:"compare_url"`
	ContentsURL      string      `json:"contents_url"`
	ContributorsURL  string      `json:"contributors_url"`
	DeploymentsURL   string      `json:"deployments_url"`
	DownloadsURL     string      `json:"downloads_url"`
	EventsURL        string      `json:"events_url"`
	ForksURL         string      `json:"forks_url"`
	GitCommitsURL    string      `json:"git_commits_url"`
	GitRefsURL       string      `json:"git_refs_url"`
	GitTagsURL       string      `json:"git_tags_url"`
	GitURL           string      `json:"git_url"`
	IssueCommentURL  string      `json:"issue_comment_url"`
	IssueEventsURL   string      `json:"issue_events_url"`
	IssuesURL        string      `json:"issues_url"`
	KeysURL          string      `json:"keys_url"`
	LabelsURL        string      `json:"labels_url"`
	LanguagesURL     string      `json:"languages_url"`
	MergesURL        string      `json:"merges_url"`
	MilestonesURL    string      `json:"milestones_url"`
	NotificationsURL string      `json:"notifications_url"`
	PullsURL         string      `json:"pulls_url"`
	ReleasesURL      string      `json:"releases_url"`
	SSHURL           string      `json:"ssh_url"`
	StargazersURL    string      `json:"stargazers_url"`
	StatusesURL      string      `json:"statuses_url"`
	SubscribersURL   string      `json:"subscribers_url"`
	SubscriptionURL  string      `json:"subscription_url"`
	TagsURL          string      `json:"tags_url"`
	TeamsURL         string      `json:"teams_url"`
	TreesURL         string      `json:"trees_url"`
	CloneURL         string      `json:"clone_url"`
	MirrorURL        string      `json:"mirror_url"`
	HooksURL         string      `json:"hooks_url"`
	SvnURL           string      `json:"svn_url"`
	Homepage         string      `json:"homepage"`
	Language         interface{} `json:"language"`
	ForksCount       int         `json:"forks_count"`
	StargazersCount  int         `json:"stargazers_count"`
	WatchersCount    int         `json:"watchers_count"`
	Size             int         `json:"size"`
	DefaultBranch    string      `json:"default_branch"`
	OpenIssuesCount  int         `json:"open_issues_count"`
	IsTemplate       bool        `json:"is_template"`
	Topics           []string    `json:"topics"`
	HasIssues        bool        `json:"has_issues"`
	HasProjects      bool        `json:"has_projects"`
	HasWiki          bool        `json:"has_wiki"`
	HasPages         bool        `json:"has_pages"`
	HasDownloads     bool        `json:"has_downloads"`
	Archived         bool        `json:"archived"`
	Disabled         bool        `json:"disabled"`
	Visibility       string      `json:"visibility"`
	PushedAt         time.Time   `json:"pushed_at"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	Permissions      struct {
		Admin bool `json:"admin"`
		Push  bool `json:"push"`
		Pull  bool `json:"pull"`
	} `json:"permissions"`
	TemplateRepository struct {
		ID       int    `json:"id"`
		NodeID   string `json:"node_id"`
		Name     string `json:"name"`
		FullName string `json:"full_name"`
		Owner    struct {
			Login             string `json:"login"`
			ID                int    `json:"id"`
			NodeID            string `json:"node_id"`
			AvatarURL         string `json:"avatar_url"`
			GravatarID        string `json:"gravatar_id"`
			URL               string `json:"url"`
			HTMLURL           string `json:"html_url"`
			FollowersURL      string `json:"followers_url"`
			FollowingURL      string `json:"following_url"`
			GistsURL          string `json:"gists_url"`
			StarredURL        string `json:"starred_url"`
			SubscriptionsURL  string `json:"subscriptions_url"`
			OrganizationsURL  string `json:"organizations_url"`
			ReposURL          string `json:"repos_url"`
			EventsURL         string `json:"events_url"`
			ReceivedEventsURL string `json:"received_events_url"`
			Type              string `json:"type"`
			SiteAdmin         bool   `json:"site_admin"`
		} `json:"owner"`
		Private          bool        `json:"private"`
		HTMLURL          string      `json:"html_url"`
		Description      string      `json:"description"`
		Fork             bool        `json:"fork"`
		URL              string      `json:"url"`
		ArchiveURL       string      `json:"archive_url"`
		AssigneesURL     string      `json:"assignees_url"`
		BlobsURL         string      `json:"blobs_url"`
		BranchesURL      string      `json:"branches_url"`
		CollaboratorsURL string      `json:"collaborators_url"`
		CommentsURL      string      `json:"comments_url"`
		CommitsURL       string      `json:"commits_url"`
		CompareURL       string      `json:"compare_url"`
		ContentsURL      string      `json:"contents_url"`
		ContributorsURL  string      `json:"contributors_url"`
		DeploymentsURL   string      `json:"deployments_url"`
		DownloadsURL     string      `json:"downloads_url"`
		EventsURL        string      `json:"events_url"`
		ForksURL         string      `json:"forks_url"`
		GitCommitsURL    string      `json:"git_commits_url"`
		GitRefsURL       string      `json:"git_refs_url"`
		GitTagsURL       string      `json:"git_tags_url"`
		GitURL           string      `json:"git_url"`
		IssueCommentURL  string      `json:"issue_comment_url"`
		IssueEventsURL   string      `json:"issue_events_url"`
		IssuesURL        string      `json:"issues_url"`
		KeysURL          string      `json:"keys_url"`
		LabelsURL        string      `json:"labels_url"`
		LanguagesURL     string      `json:"languages_url"`
		MergesURL        string      `json:"merges_url"`
		MilestonesURL    string      `json:"milestones_url"`
		NotificationsURL string      `json:"notifications_url"`
		PullsURL         string      `json:"pulls_url"`
		ReleasesURL      string      `json:"releases_url"`
		SSHURL           string      `json:"ssh_url"`
		StargazersURL    string      `json:"stargazers_url"`
		StatusesURL      string      `json:"statuses_url"`
		SubscribersURL   string      `json:"subscribers_url"`
		SubscriptionURL  string      `json:"subscription_url"`
		TagsURL          string      `json:"tags_url"`
		TeamsURL         string      `json:"teams_url"`
		TreesURL         string      `json:"trees_url"`
		CloneURL         string      `json:"clone_url"`
		MirrorURL        string      `json:"mirror_url"`
		HooksURL         string      `json:"hooks_url"`
		SvnURL           string      `json:"svn_url"`
		Homepage         string      `json:"homepage"`
		Language         interface{} `json:"language"`
		Forks            int         `json:"forks"`
		ForksCount       int         `json:"forks_count"`
		StargazersCount  int         `json:"stargazers_count"`
		WatchersCount    int         `json:"watchers_count"`
		Watchers         int         `json:"watchers"`
		Size             int         `json:"size"`
		DefaultBranch    string      `json:"default_branch"`
		OpenIssues       int         `json:"open_issues"`
		OpenIssuesCount  int         `json:"open_issues_count"`
		IsTemplate       bool        `json:"is_template"`
		License          struct {
			Key     string `json:"key"`
			Name    string `json:"name"`
			URL     string `json:"url"`
			SpdxID  string `json:"spdx_id"`
			NodeID  string `json:"node_id"`
			HTMLURL string `json:"html_url"`
		} `json:"license"`
		Topics       []string  `json:"topics"`
		HasIssues    bool      `json:"has_issues"`
		HasProjects  bool      `json:"has_projects"`
		HasWiki      bool      `json:"has_wiki"`
		HasPages     bool      `json:"has_pages"`
		HasDownloads bool      `json:"has_downloads"`
		Archived     bool      `json:"archived"`
		Disabled     bool      `json:"disabled"`
		Visibility   string    `json:"visibility"`
		PushedAt     time.Time `json:"pushed_at"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Permissions  struct {
			Admin bool `json:"admin"`
			Push  bool `json:"push"`
			Pull  bool `json:"pull"`
		} `json:"permissions"`
		AllowRebaseMerge    bool   `json:"allow_rebase_merge"`
		TempCloneToken      string `json:"temp_clone_token"`
		AllowSquashMerge    bool   `json:"allow_squash_merge"`
		AllowAutoMerge      bool   `json:"allow_auto_merge"`
		DeleteBranchOnMerge bool   `json:"delete_branch_on_merge"`
		AllowMergeCommit    bool   `json:"allow_merge_commit"`
		SubscribersCount    int    `json:"subscribers_count"`
		NetworkCount        int    `json:"network_count"`
	} `json:"template_repository"`
}

type GitHubForbbidenResponse struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
}