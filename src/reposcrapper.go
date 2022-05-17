package main

import "Sburris/reposcrapper/src/cmd"

// Settings to be moved into settings file
// settings file location = "~/.reposcrapper/settings.json"
// TODO: comma seperated list

// const DISPLAY_FORMAT_TIME_ONLY string = "15:04:05"
// const DISPLAY_FORMAT_DATE_ONLY string = ""
// const DISPLAY_FORMAT_DATE_AND_TIME_SORTABLE string = "2021-08-31 15:04:05"
// const DISPLAY_FORMAT_DATE_AND_TIME_HUMAN_READABLE string = "31 Aug 2021 15:04:05"

func main() {
	cmd.Execute()
}

// func (data *Data) DownloadFiles(pattern string) {
// 	// Search all repos
// 	regexPattern, err := regexp.Compile(pattern)
// 	checkError(err, "Creating RegEx Pattern")

// 	found := 0
// 	//repoCount := len(repos)
// 	for _, repo := range data.repos {
// 		//Log(fmt.Sprintf("[%d/%d] Search Repo: %s", rIndex, repoCount, repo.FullName))
// 		for _, leaf := range repo.Files {
// 			if(leaf.Type == "blob") {
// 				filename := filepath.Base(leaf.Path)
// 				if regexPattern.MatchString(filename) {
// 					found++
// 					Log(fmt.Sprintf("Found Match: %s/%s", repo.FullName, leaf.Path))
// 					//continue
// 					fullFilename := filepath.Join(DEFAULT_CACHE_LOCATION, repo.FullName, leaf.Path)

// 					// Get file
// 					fileBytes, err := data.gitapi.GetFile(repo, *leaf)
// 					for fileBytes == nil && err == nil {
// 						fileBytes, err = data.gitapi.GetFile(repo, *leaf)
// 					}
	
// 					SaveFile(fileBytes, fullFilename)
// 				}
// 			}
// 		}
// 	}
// 	Log(fmt.Sprintf("Found: %d", found))
// }





// func (data *Data) GetRepoTrees() {
	
// 	repoTotal := len(data.repos)
// 	index := 1
// 	for _, repo := range data.repos {
// 		treeResponse := data.gitapi.GetGithubTree(repo.Name, repo.DefaultBranch, true, index, repoTotal)
// 		for treeResponse == nil {
// 			index++
// 			treeResponse = data.gitapi.GetGithubTree(repo.Name, repo.DefaultBranch, true, index, repoTotal)
// 		}

// 		if treeResponse.Truncated {
// 			Log(fmt.Sprintf("TREE Truncated: %s", repo.Name))
// 			repo.WasTreeTruncated = true
// 			continue // don't save the files, currently.  The two projects that do this are new and look to be test projects, plus I don't want the extra 30MB need to house all of their dependencies
// 			// TODO: possibly put in exclusion folders to not add stuff to our records, this npm and other folders would be good for that.
// 		}
// 		repo.WasTreeTruncated = false

// 		//repo.Files = make([]RepositoryFile, 0, len(treeResponse.Tree))
// 		for _, leaf := range treeResponse.Tree {

// 			newLeaf := RepositoryFile {
// 				Path: leaf.Path,
// 				Mode: leaf.Mode,
// 				Type: leaf.Type,
// 				Sha: leaf.Sha,
// 				Size: leaf.Size,
// 				Url: leaf.URL,
// 			}
// 			//Log(fmt.Sprintf("New Leave Added: [%s] %s", repo.Name, newLeaf.Path))
// 			repo.Files = append(repo.Files, &newLeaf)
// 			//repos[i].Files = append(repos[i].Files, newLeaf)
// 		}
// 	}
// }

// func createRepoMap(in []Repository) (out map[string]*Repository) {
// 	for i:= 0; i < len(in); i++ {
// 		out[in[i].FullName] = &in[i]
// 	}
// 	return
// }
