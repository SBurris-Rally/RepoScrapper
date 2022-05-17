package data

import (
    "encoding/json"
    "testing"

    "github.com/stretchr/testify/assert"
)

func Test_snykService_extractReposFromProjectsList(t *testing.T) {
    inputJSON := `{
        "org": {
            "name": "Rally SCA",
            "id": "9bf47dc4-8c9d-4ee5-af4d-4270a19dd1c6"
        },
        "projects": [
            {
                "id": "08e9ce48-7401-4f51-8f6f-4699eff3b9d5",
                "name": "AudaxHealthInc/arcade-mobile-edge:build.sbt",
                "created": "2020-07-28T21:40:49.971Z",
                "origin": "github",
                "type": "sbt",
                "readOnly": false,
                "testFrequency": "daily",
                "isMonitored": true,
                "totalDependencies": 1,
                "issueCountsBySeverity": {
                    "low": 0,
                    "high": 0,
                    "medium": 0,
                    "critical": 0
                },
                "imageId": "",
                "imageTag": "1.0.0-SNAPSHOT",
                "imagePlatform": "",
                "imageBaseImage": "",
                "lastTestedDate": "2022-02-22T19:21:44.550Z",
                "browseUrl": "https://app.snyk.io/org/rally-sca-poc/project/08e9ce48-7401-4f51-8f6f-4699eff3b9d5",
                "owner": null,
                "importingUser": {
                    "id": "6b1ab814-ffff-44d9-ac07-b7f777a6aa5b",
                    "name": "Micky Mouse",
                    "username": "Micky Mouse",
                    "email": "micky.mouse@rallyhealth.com"
                },
                "tags": [],
                "attributes": {
                    "criticality": [],
                    "lifecycle": [],
                    "environment": []
                },
                "branch": "master"
            }
        ]
    }`
    var input SnykProjectsListPayload
    json.Unmarshal([]byte(inputJSON), &input)
    expected := []string {"AudaxHealthInc/arcade-mobile-edge"}

    output := extractReposFromProjectslist(input)

    assert.EqualValues(t, expected, output)
}