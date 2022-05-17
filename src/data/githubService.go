package data

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"encoding/json"
	"io/ioutil"
	"net/http"

	utils "Sburris/reposcrapper/src/utils"
)

//const DEFAULT_GITHUB_OWNER string = "AudaxHealthInc"
//const GITHUB_API_BASE string = "https://api.github.com/"

type GithubAPI struct {
	APIBase string
	TokenEnvKey string
	Token string
}

func CreateGithubService(apiBase string, tokenEnvKey string) *GithubAPI {
	return &GithubAPI{
		APIBase: apiBase,
		TokenEnvKey: tokenEnvKey,
		Token: os.Getenv(tokenEnvKey),
	}
}

func CreateGithubDefaultService() *GithubAPI {
	return &GithubAPI{
		APIBase: GitHubAPIBase,
		TokenEnvKey: GitHubAPITokenEnv,
		Token: os.Getenv(GitHubAPITokenEnv),
	}
}

func (github *GithubAPI) GetFile(repo *Repository, file RepositoryFile) ([]byte, error) {
	fileURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", DefaultGitHubOwner, repo.Name, file.Path)

	client := &http.Client{}
	fileReq, err := http.NewRequest("GET", fileURL, nil)
	utils.CheckError(err, "New Request for GET user repo")

	fileReq.Header.Add("Accept", "application/vnd.github.VERSION.raw")
	fileReq.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", github.Token))

	fileResp, err := client.Do(fileReq)
	utils.CheckError(err, "client.Do")
	
	defer fileResp.Body.Close()
	fileBytes, err := ioutil.ReadAll(fileResp.Body)
	utils.CheckError(err, "Ready repo response body")

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
		return nil, fmt.Errorf("%d", fileResp.StatusCode)
	}

	return fileBytes, nil
}

func (github GithubAPI) GetAllRepos() (map[string]*Repository, error) {
	// TODO: Is the error response necessary?
	// TODO: figure out a better way then hardcoding repository capacity
	allRepos := make(map[string]*Repository)

	count := 0
	// TODO: replace with while loop?
	for { // Loop until done
		count++
		repos, err := github.GetRepos(count)
		
		// TODO: Why is this here... what was the reason?
		for err != nil {
			repos, err = github.GetRepos(count)
		}

		if repos != nil {
			// TODO: this is where the update action needs to be determined
			for _, repo := range repos {
				allRepos[repo.Name] = repo
			}
		} else {
			break // no records, implies we are at the end of the repos
		}
	}

	return allRepos, nil
}

func (github GithubAPI) GetRepos(page int) (out []*Repository, err error) {
	utils.Log(fmt.Sprintf("Performing Http Get - Page %d", page))

	client := &http.Client{}
	// https://api.github.com/search/repositories?q=user:sburris fork:true&per_page=100   // TODO: why is this rate-limit at 60?
	// https://api.github.com/users/Sburris/repos?per_page=100
	// req, err := http.NewRequest("GET", "https://api.github.com/search/repositories?q=user:sburris%20fork:true&per_page=100", nil)
	// TODO: Currently only getting one page, need to refactor to grab all
	url := fmt.Sprintf("https://api.github.com/orgs/AudaxHealthInc/repos?per_page=100&page=%d", page)
	req, err := http.NewRequest("GET", url, nil)
	utils.CheckError(err, "New Request for GET user repo")

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", github.Token))

	//Log("Performing Call")
	resp, err := client.Do(req)
	utils.CheckError(err, "client.Do")
	
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	utils.CheckError(err, "Ready repo response body")
	
	if resp.StatusCode != 200 {
		PrintResponseHeaders(*resp)
		SleepUntil(resp.Header["X-Ratelimit-Reset"][0])
		return nil, errors.New("Timeout")
	}

	var repoResponse GithubRepoListGet
	json.Unmarshal(bodyBytes, &repoResponse)

	for _, repo := range repoResponse {
		newRepo := Repository {
			ID: repo.ID,
			NodeID: repo.NodeID,
			Name: repo.Name,
			FullName: repo.FullName,
			IsPrivate: repo.Private,
			IsFork: repo.Fork,
			//IsArchived: repo.Archived,
			HTMLURL: repo.HTMLURL,
			Description: repo.Description,
			CreatedAt: repo.CreatedAt,
			UpdatedAt: repo.UpdatedAt,
			PushedAt: repo.PushedAt,
			DefaultBranch: repo.DefaultBranch,
			SSHURL: repo.SSHURL,
		}
		//newRepo.Files = make([]RepositoryFile, 0, 200)
		out = append(out, &newRepo)
	}

	return out, nil
}

func (github GithubAPI) GetPullRequests(repo string) (GithubPullRequestResponse, error) {
	utils.Log(fmt.Sprintf("Getting Pull Requests for Repo '%s'", repo))

	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/repos/AudaxHealthInc/%s/pulls?state=all&per_page=100", repo)
	req, err := http.NewRequest("GET", url, nil)
	utils.CheckError(err, "New Request for GET pull requests")

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", github.Token))

	//Log("Performing Call")
	resp, err := client.Do(req)
	utils.CheckError(err, "client.Do")
	
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	utils.CheckError(err, "Ready repo response body")
	
	if resp.StatusCode != 200 {
		PrintResponseHeaders(*resp)
		SleepUntil(resp.Header["X-Ratelimit-Reset"][0])
		return nil, errors.New("Timeout")
	}

	var repoResponse GithubPullRequestResponse
	json.Unmarshal(bodyBytes, &repoResponse)

	// pullRequests := make([]string, 0)
	// for _, pullRequest := range repoResponse {
	// 	pullRequests = append(pullRequests, pullRequest.Title)
	// }

	return repoResponse, nil
}

func (github *GithubAPI) GetGithubTree(repoName string, defaultBranch string, isRecursive bool, index int, total int) *GithubTreeResponse {
	client := &http.Client{}
	treeRequestURL := fmt.Sprintf("https://api.github.com/repos/AudaxHealthInc/%s/git/trees/%s", repoName, defaultBranch)
	if isRecursive {
		treeRequestURL = treeRequestURL + "?recursive=1"
	}

	//treeRequestUrl := "https://api.github.com/repos/SBurris/SecurityNotes/git/trees/master?recursive=1"
	utils.Log(fmt.Sprintf("[%d/%d]Retrieving Tree for Repo: %s (%s)",index, total, repoName, treeRequestURL))
	treeReq, treeErr := http.NewRequest("GET", treeRequestURL, nil)
	utils.CheckError(treeErr, "New Request for GET user repo")

	treeReq.Header.Add("Accept", "application/json")
	treeReq.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", github.Token))

	treeResp, err := client.Do(treeReq)
	utils.CheckError(err, "client.Do")
	fmt.Printf("X-RateLimit-Remaining: %s\n", treeResp.Header["X-Ratelimit-Remaining"][0])

	defer treeResp.Body.Close()
	treeBodyBytes, err := ioutil.ReadAll(treeResp.Body)
	utils.CheckError(err, "Ready repo response body")

	if treeResp.StatusCode != 200 && treeResp.StatusCode != 404 && treeResp.StatusCode != 409 {
		PrintResponseHeaders(*treeResp)
		SleepUntil(treeResp.Header["X-Ratelimit-Reset"][0])
		return nil
	}

	var treeResponse GithubTreeResponse
	json.Unmarshal(treeBodyBytes, &treeResponse)

	return &treeResponse
}

func PrintResponseHeaders(resp http.Response) {
	fmt.Printf("Response Status Code: %d\n", resp.StatusCode)
	fmt.Printf("X-RateLimit-Limit: %s\n", resp.Header["X-Ratelimit-Limit"][0])
	fmt.Printf("X-RateLimit-Remaining: %s\n", resp.Header["X-Ratelimit-Remaining"][0])
	fmt.Printf("X-RateLimit-Used: %s\n", resp.Header["X-Ratelimit-Used"][0])
	fmt.Printf("X-RateLimit-Reset: %s\n", resp.Header["X-Ratelimit-Reset"][0])
}

func SleepUntil(reset string) {
	expireTime, _ := strconv.Atoi(reset)
	resetTime := time.Unix(int64(expireTime), 0).Add(time.Second * 30)
	waitPeriod := resetTime.Unix() - time.Now().Unix()

	dur := time.Until(resetTime)
	utils.Log(fmt.Sprintf("Sleeping For %s [%d]", dur.String(), waitPeriod))
	time.Sleep(time.Until(resetTime))
}


// GithubSearchResponse is the response object for GitHub's SearchAPI endpoint
// Source: https://mholt.github.io/json-to-go/
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

// GithubRepoListGet is the ResponseObject for GitHub's repo endpoint
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



type GithubPullRequestResponse []struct {
	URL               string `json:"url"`
	ID                int    `json:"id"`
	NodeID            string `json:"node_id"`
	HTMLURL           string `json:"html_url"`
	DiffURL           string `json:"diff_url"`
	PatchURL          string `json:"patch_url"`
	IssueURL          string `json:"issue_url"`
	CommitsURL        string `json:"commits_url"`
	ReviewCommentsURL string `json:"review_comments_url"`
	ReviewCommentURL  string `json:"review_comment_url"`
	CommentsURL       string `json:"comments_url"`
	StatusesURL       string `json:"statuses_url"`
	Number            int    `json:"number"`
	State             string `json:"state"`
	Locked            bool   `json:"locked"`
	Title             string `json:"title"`
	User              struct {
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
	} `json:"user"`
	Body   string `json:"body"`
	Labels []struct {
		ID          int    `json:"id"`
		NodeID      string `json:"node_id"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Default     bool   `json:"default"`
	} `json:"labels"`
	Milestone struct {
		URL         string `json:"url"`
		HTMLURL     string `json:"html_url"`
		LabelsURL   string `json:"labels_url"`
		ID          int    `json:"id"`
		NodeID      string `json:"node_id"`
		Number      int    `json:"number"`
		State       string `json:"state"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Creator     struct {
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
		} `json:"creator"`
		OpenIssues   int       `json:"open_issues"`
		ClosedIssues int       `json:"closed_issues"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		ClosedAt     time.Time `json:"closed_at"`
		DueOn        time.Time `json:"due_on"`
	} `json:"milestone"`
	ActiveLockReason string    `json:"active_lock_reason"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ClosedAt         time.Time `json:"closed_at"`
	MergedAt         time.Time `json:"merged_at"`
	MergeCommitSha   string    `json:"merge_commit_sha"`
	Assignee         struct {
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
	} `json:"assignee"`
	Assignees []struct {
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
	} `json:"assignees"`
	RequestedReviewers []struct {
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
	} `json:"requested_reviewers"`
	RequestedTeams []struct {
		ID              int         `json:"id"`
		NodeID          string      `json:"node_id"`
		URL             string      `json:"url"`
		HTMLURL         string      `json:"html_url"`
		Name            string      `json:"name"`
		Slug            string      `json:"slug"`
		Description     string      `json:"description"`
		Privacy         string      `json:"privacy"`
		Permission      string      `json:"permission"`
		MembersURL      string      `json:"members_url"`
		RepositoriesURL string      `json:"repositories_url"`
		Parent          interface{} `json:"parent"`
	} `json:"requested_teams"`
	Head struct {
		Label string `json:"label"`
		Ref   string `json:"ref"`
		Sha   string `json:"sha"`
		User  struct {
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
		} `json:"user"`
		Repo struct {
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
			AllowRebaseMerge    bool        `json:"allow_rebase_merge"`
			TemplateRepository  interface{} `json:"template_repository"`
			TempCloneToken      string      `json:"temp_clone_token"`
			AllowSquashMerge    bool        `json:"allow_squash_merge"`
			AllowAutoMerge      bool        `json:"allow_auto_merge"`
			DeleteBranchOnMerge bool        `json:"delete_branch_on_merge"`
			AllowMergeCommit    bool        `json:"allow_merge_commit"`
			SubscribersCount    int         `json:"subscribers_count"`
			NetworkCount        int         `json:"network_count"`
			License             struct {
				Key     string `json:"key"`
				Name    string `json:"name"`
				URL     string `json:"url"`
				SpdxID  string `json:"spdx_id"`
				NodeID  string `json:"node_id"`
				HTMLURL string `json:"html_url"`
			} `json:"license"`
			Forks      int `json:"forks"`
			OpenIssues int `json:"open_issues"`
			Watchers   int `json:"watchers"`
		} `json:"repo"`
	} `json:"head"`
	Base struct {
		Label string `json:"label"`
		Ref   string `json:"ref"`
		Sha   string `json:"sha"`
		User  struct {
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
		} `json:"user"`
		Repo struct {
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
			AllowRebaseMerge    bool        `json:"allow_rebase_merge"`
			TemplateRepository  interface{} `json:"template_repository"`
			TempCloneToken      string      `json:"temp_clone_token"`
			AllowSquashMerge    bool        `json:"allow_squash_merge"`
			AllowAutoMerge      bool        `json:"allow_auto_merge"`
			DeleteBranchOnMerge bool        `json:"delete_branch_on_merge"`
			AllowMergeCommit    bool        `json:"allow_merge_commit"`
			SubscribersCount    int         `json:"subscribers_count"`
			NetworkCount        int         `json:"network_count"`
			License             struct {
				Key     string `json:"key"`
				Name    string `json:"name"`
				URL     string `json:"url"`
				SpdxID  string `json:"spdx_id"`
				NodeID  string `json:"node_id"`
				HTMLURL string `json:"html_url"`
			} `json:"license"`
			Forks      int `json:"forks"`
			OpenIssues int `json:"open_issues"`
			Watchers   int `json:"watchers"`
		} `json:"repo"`
	} `json:"base"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
		HTML struct {
			Href string `json:"href"`
		} `json:"html"`
		Issue struct {
			Href string `json:"href"`
		} `json:"issue"`
		Comments struct {
			Href string `json:"href"`
		} `json:"comments"`
		ReviewComments struct {
			Href string `json:"href"`
		} `json:"review_comments"`
		ReviewComment struct {
			Href string `json:"href"`
		} `json:"review_comment"`
		Commits struct {
			Href string `json:"href"`
		} `json:"commits"`
		Statuses struct {
			Href string `json:"href"`
		} `json:"statuses"`
	} `json:"_links"`
	AuthorAssociation string      `json:"author_association"`
	AutoMerge         interface{} `json:"auto_merge"`
	Draft             bool        `json:"draft"`
}