/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"encoding/json"
	"fmt"
	"regexp"
	"runtime/debug"

	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitee/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ExtractApiIssuesMeta = core.SubTaskMeta{
	Name:             "extractApiIssues",
	EntryPoint:       ExtractApiIssues,
	EnabledByDefault: true,
	Description:      "Extract raw Issues data into tool layer table gitee_issues",
}

type IssuesResponse struct {
	GiteeId       int    `json:"id"`
	Url           string `json:"url"`
	RepositoryUrl string `json:"repository_url"`
	Number        string `json:"number"`
	State         string `json:"state"`
	Title         string
	Body          string
	HtmlUrl       string `json:"html_url"`
	CommentsUrl   string `json:"comments_url"`
	PullRequest   struct {
		Url     string `json:"url"`
		HtmlUrl string `json:"html_url"`
	} `json:"pull_request"`
	Labels []struct {
		Id           int
		RepositoryId int                `json:"repository_id"`
		Name         string             `json:"name"`
		CreatedAt    helper.Iso8601Time `json:"created_at"`
		UpdatedAt    helper.Iso8601Time `json:"updated_at"`
	} `json:"labels"`
	Repository struct {
		Id       int
		FullName string `json:"full_name"`
		Url      string `json:"url"`
	} `json:"repository"`
	Assignee *struct {
		Login string
		Id    int
	}
	User *struct {
		Login string
		Id    int
		Name  string
	}
	Comments        int                 `json:"comments"`
	Priority        int                 `json:"priority"`
	IssueType       string              `json:"issue_type"`
	SecurityHole    bool                `json:"security_hole"`
	IssueState      string              `json:"issue_state"`
	Branch          string              `json:"branch"`
	FinishAt        *helper.Iso8601Time `json:"finished_at"`
	GiteeCreatedAt  helper.Iso8601Time  `json:"created_at"`
	GiteeUpdatedAt  helper.Iso8601Time  `json:"updated_at"`
	IssueTypeDetail struct {
		Id        int
		Title     string
		Ident     string
		CreatedAt helper.Iso8601Time `json:"created_at"`
		UpdatedAt helper.Iso8601Time `json:"updated_at"`
	}
	IssueStateDetail struct {
		Id        int
		Title     string
		Serial    string
		CreatedAt helper.Iso8601Time `json:"created_at"`
		UpdatedAt helper.Iso8601Time `json:"updated_at"`
	}
}

func ExtractApiIssues(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_ISSUE_TABLE)
	config := data.Options.TransformationRules
	var issueSeverityRegex *regexp.Regexp
	var issueComponentRegex *regexp.Regexp
	var issuePriorityRegex *regexp.Regexp
	var issueTypeBugRegex *regexp.Regexp
	var issueTypeRequirementRegex *regexp.Regexp
	var issueTypeIncidentRegex *regexp.Regexp
	var issueSeverity = config.IssueSeverity
	var err error
	if len(issueSeverity) > 0 {
		issueSeverityRegex, err = regexp.Compile(issueSeverity)
		if err != nil {
			return fmt.Errorf("regexp Compile issueSeverity failed:[%s] stack:[%s]", err.Error(), debug.Stack())
		}
	}
	var issueComponent = config.IssueComponent
	if len(issueComponent) > 0 {
		issueComponentRegex, err = regexp.Compile(issueComponent)
		if err != nil {
			return fmt.Errorf("regexp Compile issueComponent failed:[%s] stack:[%s]", err.Error(), debug.Stack())
		}
	}
	var issuePriority = config.IssuePriority
	if len(issuePriority) > 0 {
		issuePriorityRegex, err = regexp.Compile(issuePriority)
		if err != nil {
			return fmt.Errorf("regexp Compile issuePriority failed:[%s] stack:[%s]", err.Error(), debug.Stack())
		}
	}
	var issueTypeBug = config.IssueTypeBug
	if len(issueTypeBug) > 0 {
		issueTypeBugRegex, err = regexp.Compile(issueTypeBug)
		if err != nil {
			return fmt.Errorf("regexp Compile issueTypeBug failed:[%s] stack:[%s]", err.Error(), debug.Stack())
		}
	}
	var issueTypeRequirement = config.IssueTypeRequirement
	if len(issueTypeRequirement) > 0 {
		issueTypeRequirementRegex, err = regexp.Compile(issueTypeRequirement)
		if err != nil {
			return fmt.Errorf("regexp Compile issueTypeRequirement failed:[%s] stack:[%s]", err.Error(), debug.Stack())
		}
	}
	var issueTypeIncident = config.IssueTypeIncident
	if len(issueTypeIncident) > 0 {
		issueTypeIncidentRegex, err = regexp.Compile(issueTypeIncident)
		if err != nil {
			return fmt.Errorf("regexp Compile issueTypeIncident failed:[%s] stack:[%s]", err.Error(), debug.Stack())
		}
	}

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			body := &IssuesResponse{}
			err := json.Unmarshal(row.Data, body)
			if err != nil {
				return nil, err
			}
			// need to extract 2 kinds of entities here
			if body.GiteeId == 0 {
				return nil, nil
			}
			//If this is a pr, ignore
			if body.PullRequest.Url != "" {
				return nil, nil
			}
			results := make([]interface{}, 0, 2)
			giteeIssue, err := convertGiteeIssue(body, data.Options.ConnectionId, data.Repo.GiteeId)
			if err != nil {
				return nil, err
			}
			for _, label := range body.Labels {
				results = append(results, &models.GiteeIssueLabel{
					ConnectionId: data.Options.ConnectionId,
					IssueId:      giteeIssue.GiteeId,
					LabelName:    label.Name,
				})
				if issueSeverityRegex != nil {
					groups := issueSeverityRegex.FindStringSubmatch(label.Name)
					if len(groups) > 0 {
						giteeIssue.Severity = groups[1]
					}
				}

				if issueComponentRegex != nil {
					groups := issueComponentRegex.FindStringSubmatch(label.Name)
					if len(groups) > 0 {
						giteeIssue.Component = groups[1]
					}
				}

				if issuePriorityRegex != nil {
					groups := issuePriorityRegex.FindStringSubmatch(label.Name)
					if len(groups) > 0 {
						giteeIssue.Priority = groups[1]
					}
				}

				if issueTypeBugRegex != nil {
					if ok := issueTypeBugRegex.MatchString(label.Name); ok {
						giteeIssue.Type = ticket.BUG
					}
				}

				if issueTypeRequirementRegex != nil {
					if ok := issueTypeRequirementRegex.MatchString(label.Name); ok {
						giteeIssue.Type = ticket.REQUIREMENT
					}
				}

				if issueTypeIncidentRegex != nil {
					if ok := issueTypeIncidentRegex.MatchString(label.Name); ok {
						giteeIssue.Type = ticket.INCIDENT
					}
				}
			}
			results = append(results, giteeIssue)

			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}
func convertGiteeIssue(issue *IssuesResponse, connectionId uint64, repositoryId int) (*models.GiteeIssue, error) {
	giteeIssue := &models.GiteeIssue{
		ConnectionId:   connectionId,
		GiteeId:        issue.GiteeId,
		RepoId:         repositoryId,
		Number:         issue.Number,
		State:          issue.State,
		Title:          issue.Title,
		Body:           issue.Body,
		Url:            issue.HtmlUrl,
		ClosedAt:       helper.Iso8601TimeToTime(issue.FinishAt),
		GiteeCreatedAt: issue.GiteeCreatedAt.ToTime(),
		GiteeUpdatedAt: issue.GiteeUpdatedAt.ToTime(),
	}

	if issue.Assignee != nil {
		giteeIssue.AssigneeId = issue.Assignee.Id
		giteeIssue.AssigneeName = issue.Assignee.Login
	}
	if issue.User != nil {
		giteeIssue.AuthorId = issue.User.Id
		giteeIssue.AuthorName = issue.User.Login
	}
	if issue.FinishAt != nil {
		giteeIssue.LeadTimeMinutes = uint(issue.FinishAt.ToTime().Sub(issue.GiteeCreatedAt.ToTime()).Minutes())
	}

	return giteeIssue, nil
}
