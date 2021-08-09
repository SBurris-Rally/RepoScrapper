package main

import (
	"fmt"
	"os"
	"time"
	//"log" 			 TODO: Look ingo
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Settings to be moved into settings file
// settings file location = "~/.reposcrapper/settings.json"
const DEFAULT_GITHUB_OWNER string = "Sburris"
const DEFAULT_GITHUB_TYPE string = "user" // user or org
const CACHE_LOCATION string = "~/.reposcrapper/cache/"

const DISPLAY_FORMAT_TIME_ONLY_HIGH_PERCISION string = "15:04:05.000"
// const DISPLAY_FORMAT_TIME_ONLY string = "15:04:05"
// const DISPLAY_FORMAT_DATE_ONLY string = ""
// const DISPLAY_FORMAT_DATE_AND_TIME_SORTABLE string = "2021-08-31 15:04:05"
// const DISPLAY_FORMAT_DATE_AND_TIME_HUMAN_READABLE string = "31 Aug 2021 15:04:05"

const GITHUB_API_BASE string = "https://api.github.com/"
const GITHUB_API_TOKEN_ENV string = "GITHUB_TOKEN"

var GITHUB_API_TOKEN = "EMPTY"

func main() {
	GITHUB_API_TOKEN = os.Getenv(GITHUB_API_TOKEN_ENV)

	Log("Performing Http Get")
	client := &http.Client{}
	// https://api.github.com/search/repositories?q=user:sburris fork:true&per_page=100   // TODO: why is this rate-limit at 60?
	// https://api.github.com/users/Sburris/repos?per_page=100
	req, err := http.NewRequest("GET", "https://api.github.com/search/repositories?q=user:sburris%20fork:true&per_page=100", nil)
	checkError(err, "New Request for GET user repo")

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", GITHUB_API_TOKEN))
	fmt.Printf("Auth Header: %s\n", req.Header["Authorization"])

	Log("Performing Call")
	resp, err := client.Do(req)
	checkError(err, "client.Do")
	
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	checkError(err, "Ready repo response body")

	var repoResponse GithubSearchResponse
	json.Unmarshal(bodyBytes, &repoResponse)

	fmt.Printf("Response Status Code: %d\n", resp.StatusCode)
	fmt.Printf("X-RateLimit-Limit: %s\n", resp.Header["X-Ratelimit-Limit"][0])
	fmt.Printf("X-RateLimit-Remaining: %s\n", resp.Header["X-Ratelimit-Remaining"][0])
	fmt.Printf("X-RateLimit-Used: %s\n", resp.Header["X-Ratelimit-Used"][0])
	fmt.Printf("X-RateLimit-Reset: %s\n", resp.Header["X-Ratelimit-Reset"][0])
	fmt.Printf("Now %d\n", time.Now().Unix())

	for i, repo := range repoResponse.Items {
		fmt.Printf("Repo[%d]: %s\n", i, repo.FullName)
	}

	repos := make([]Repository,repoResponse.TotalCount)
	for i, repo := range repoResponse.Items {
		newRepo := Repository {
			Id: repo.ID,
			NodeId: repo.NodeID,
			Name: repo.Name,
			FullName: repo.FullName,
			IsPrivate: repo.Private,
			IsFork: repo.Fork,
			HtmlUrl: repo.HTMLURL,
			Description: repo.Description,
			CreatedAt: repo.CreatedAt,
			UpdatedAt: repo.UpdatedAt,
			PushedAt: repo.PushedAt,
			DefaultBranch: repo.DefaultBranch,
		}

		Log(fmt.Sprintf("Retrieving Tree for Repo: %s\n", repo.Name))
		treeClient := &http.Client{}
		treeReq, treeErr := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/sburris/%s/git/trees/master?recursive=1", repo.Name), nil)
		checkError(treeErr, "New Request for GET user repo")

		treeReq.Header.Add("Accept", "application/json")
		treeReq.Header.Add("Authorization", fmt.Sprintf("%s %s", "Token", GITHUB_API_TOKEN))

		Log("Performing Call Repo Call")
		resp, err := treeClient.Do(req)
		checkError(err, "client.Do")
		
		defer resp.Body.Close()
		treeBodyBytes, err := ioutil.ReadAll(resp.Body)
		checkError(err, "Ready repo response body")

		var treeResponse GithubTreeResponse
		json.Unmarshal(treeBodyBytes, &treeResponse)

		newRepo.Files := make([]Repository.Files,len(treeResponse.Tree))
		for x, leaf := range treeResponse.Tree {
			newLeaf := Repository.Files {
				Path
			}

			newRepo.Files[x] = newLeaf
		}

		repos[i] = newRepo
	
	}

}

func Log(msg string) {
	fmt.Printf("[%s] %s", time.Now().Format(DISPLAY_FORMAT_TIME_ONLY_HIGH_PERCISION), msg)
}

func checkError(err error, msg string) {
	if err != nil {
		Log(fmt.Sprintf("[ERROR] %s", msg))
		os.Exit(1)
	}
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
	Files []struct {
		Path string
		Mode string
		Type string
		Sha string
		Size int
		Url string
	}
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
