// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ghutil_test

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-github/github"

	"github.com/google/code-review-bot/ghutil"
)

type MockGitHubClient struct {
	Organizations *ghutil.MockOrganizationsService
	PullRequests  *ghutil.MockPullRequestsService
	Issues        *ghutil.MockIssuesService
	Repositories  *ghutil.MockRepositoriesService
}

func NewMockGitHubClient(ghc *ghutil.GitHubClient, ctrl *gomock.Controller) *MockGitHubClient {
	mockGhc := &MockGitHubClient{
		Organizations: ghutil.NewMockOrganizationsService(ctrl),
		PullRequests:  ghutil.NewMockPullRequestsService(ctrl),
		Issues:        ghutil.NewMockIssuesService(ctrl),
		Repositories:  ghutil.NewMockRepositoriesService(ctrl),
	}

	// Patch the original GitHubClient with our mock services.
	ghc.Organizations = mockGhc.Organizations
	ghc.PullRequests = mockGhc.PullRequests
	ghc.Issues = mockGhc.Issues
	ghc.Repositories = mockGhc.Repositories

	return mockGhc
}

// Common parameters used across most, if not all, tests.
var (
	ctrl    *gomock.Controller
	ghc     *ghutil.GitHubClient
	mockGhc *MockGitHubClient

	noLabel *github.Label = nil
)

const (
	orgName   = "org"
	repoName  = "repo"
	emptyRepo = ""
)

func setUp(t *testing.T) {
	ctrl = gomock.NewController(t)
	ghc = &ghutil.GitHubClient{}
	mockGhc = NewMockGitHubClient(ghc, ctrl)
}

func tearDown(t *testing.T) {
	ctrl.Finish()
}

func TestGetAllRepos_OrgAndRepo(t *testing.T) {
	setUp(t)
	defer tearDown(t)

	repo := github.Repository{}

	mockGhc.Repositories.EXPECT().Get(orgName, repoName).Return(&repo, nil, nil)

	repos := ghc.GetAllRepos(orgName, repoName)
	if len(repos) != 1 {
		t.Logf("repos is not of length 1: %v", repos)
		t.Fail()
	}
}

func TestGetAllRepos_OrgOnly(t *testing.T) {
	setUp(t)
	defer tearDown(t)

	expectedRepos := []*github.Repository{
		&github.Repository{},
		&github.Repository{},
	}

	mockGhc.Repositories.EXPECT().List(orgName, nil).Return(expectedRepos, nil, nil)

	actualRepos := ghc.GetAllRepos(orgName, "")
	if len(expectedRepos) != len(actualRepos) {
		t.Logf("Expected repos: %v, actual repos: %v", expectedRepos, actualRepos)
		t.Fail()
	}
}

func TestVerifyRepoHasClaLabels_NoLabels(t *testing.T) {
	setUp(t)
	defer tearDown(t)

	mockGhc.Issues.EXPECT().GetLabel(orgName, repoName, ghutil.LabelClaYes).Return(noLabel, nil, nil)
	mockGhc.Issues.EXPECT().GetLabel(orgName, repoName, ghutil.LabelClaNo).Return(noLabel, nil, nil)

	if ghc.VerifyRepoHasClaLabels(orgName, repoName) {
		t.Log("Should have returned false")
		t.Fail()
	}
}

func TestVerifyRepoHasClaLabels_HasYesOnly(t *testing.T) {
	setUp(t)
	defer tearDown(t)

	label := github.Label{}

	mockGhc.Issues.EXPECT().GetLabel(orgName, repoName, ghutil.LabelClaYes).Return(&label, nil, nil)
	mockGhc.Issues.EXPECT().GetLabel(orgName, repoName, ghutil.LabelClaNo).Return(noLabel, nil, nil)

	if ghc.VerifyRepoHasClaLabels(orgName, repoName) {
		t.Log("Should have returned false")
		t.Fail()
	}
}

func TestVerifyRepoHasClaLabels_HasNoOnly(t *testing.T) {
	setUp(t)
	defer tearDown(t)

	label := github.Label{}

	mockGhc.Issues.EXPECT().GetLabel(orgName, repoName, ghutil.LabelClaYes).Return(noLabel, nil, nil)
	mockGhc.Issues.EXPECT().GetLabel(orgName, repoName, ghutil.LabelClaNo).Return(&label, nil, nil)

	if ghc.VerifyRepoHasClaLabels(orgName, repoName) {
		t.Log("Should have returned false")
		t.Fail()
	}
}

func TestVerifyRepoHasClaLabels_YesAndNoLabels(t *testing.T) {
	setUp(t)
	defer tearDown(t)

	labelYes := github.Label{}
	labelNo := github.Label{}

	mockGhc.Issues.EXPECT().GetLabel(orgName, repoName, ghutil.LabelClaYes).Return(&labelYes, nil, nil)
	mockGhc.Issues.EXPECT().GetLabel(orgName, repoName, ghutil.LabelClaNo).Return(&labelNo, nil, nil)

	if !ghc.VerifyRepoHasClaLabels(orgName, repoName) {
		t.Log("Should have returned true")
		t.Fail()
	}
}
