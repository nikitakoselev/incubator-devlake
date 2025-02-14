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

package e2e

import (
	"testing"

	"github.com/apache/incubator-devlake/helpers/e2ehelper"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/jira/impl"
	"github.com/apache/incubator-devlake/plugins/jira/models"
	"github.com/apache/incubator-devlake/plugins/jira/tasks"
)

func TestIssueDataFlow(t *testing.T) {
	var plugin impl.Jira
	dataflowTester := e2ehelper.NewDataFlowTester(t, "jira", plugin)

	taskData := &tasks.JiraTaskData{
		Options: &tasks.JiraOptions{
			ConnectionId:        2,
			BoardId:             8,
			TransformationRules: tasks.TransformationRules{StoryPointField: "customfield_10024"}},
	}

	// import raw data table
	dataflowTester.ImportCsvIntoRawTable("./raw_tables/_raw_jira_api_issues.csv", "_raw_jira_api_issues")

	// verify issue extraction
	dataflowTester.FlushTabler(&models.JiraIssue{})
	dataflowTester.FlushTabler(&models.JiraBoardIssue{})
	dataflowTester.FlushTabler(&models.JiraSprintIssue{})
	dataflowTester.FlushTabler(&models.JiraIssueChangelogs{})
	dataflowTester.FlushTabler(&models.JiraIssueChangelogItems{})
	dataflowTester.FlushTabler(&models.JiraWorklog{})
	dataflowTester.FlushTabler(&models.JiraAccount{})
	dataflowTester.Subtask(tasks.ExtractIssuesMeta, taskData)
	dataflowTester.VerifyTable(
		models.JiraIssue{},
		"./snapshot_tables/_tool_jira_issues.csv",
		[]string{
			"connection_id",
			"issue_id",
			"project_id",
			"self",
			"issue_key",
			"summary",
			"type",
			"epic_key",
			"status_name",
			"status_key",
			"story_point",
			"original_estimate_minutes",
			"aggregate_estimate_minutes",
			"remaining_estimate_minutes",
			"creator_account_id",
			"creator_account_type",
			"creator_display_name",
			"assignee_account_id",
			"assignee_account_type",
			"assignee_display_name",
			"priority_id",
			"priority_name",
			"parent_id",
			"parent_key",
			"sprint_id",
			"sprint_name",
			"resolution_date",
			"created",
			"updated",
			"spent_minutes",
			"lead_time_minutes",
			"std_story_point",
			"std_type",
			"std_status",
			"_raw_data_params",
			"_raw_data_table",
			"_raw_data_id",
			"_raw_data_remark",
			"icon_url",
		},
	)

	dataflowTester.VerifyTable(
		models.JiraBoardIssue{},
		"./snapshot_tables/_tool_jira_board_issues.csv",
		[]string{"connection_id", "board_id", "issue_id"},
	)

	// verify issue conversion
	dataflowTester.FlushTabler(&ticket.Issue{})
	dataflowTester.FlushTabler(&ticket.BoardIssue{})
	dataflowTester.Subtask(tasks.ConvertIssuesMeta, taskData)
	dataflowTester.VerifyTable(
		ticket.Issue{},
		"./snapshot_tables/issues.csv",
		[]string{
			"id",
			"url",
			"icon_url",
			"issue_key",
			"title",
			"description",
			"epic_key",
			"type",
			"status",
			"original_status",
			"story_point",
			"resolution_date",
			"created_date",
			"updated_date",
			"lead_time_minutes",
			"parent_issue_id",
			"priority",
			"original_estimate_minutes",
			"time_spent_minutes",
			"time_remaining_minutes",
			"creator_id",
			"creator_name",
			"assignee_id",
			"assignee_name",
			"severity",
			"component",
		},
	)
	dataflowTester.VerifyTable(
		ticket.BoardIssue{},
		"./snapshot_tables/board_issues.csv",
		[]string{"board_id", "issue_id"},
	)
}
