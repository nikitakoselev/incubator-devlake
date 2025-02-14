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
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
	"github.com/apache/incubator-devlake/plugins/tapd/models"
)

func ConvertIteration(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_ITERATION_TABLE, false)
	logger := taskCtx.GetLogger()
	db := taskCtx.GetDal()
	logger.Info("collect board:%d", data.Options.WorkspaceId)
	iterIdGen := didgen.NewDomainIdGenerator(&models.TapdIteration{})
	clauses := []dal.Clause{
		dal.From(&models.TapdIteration{}),
		dal.Where("connection_id = ? AND workspace_id = ?", data.Options.ConnectionId, data.Options.WorkspaceId),
	}

	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	workspaceIdGen := didgen.NewDomainIdGenerator(&models.TapdWorkspace{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		InputRowType:       reflect.TypeOf(models.TapdIteration{}),
		Input:              cursor,
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			iter := inputRow.(*models.TapdIteration)
			domainIter := &ticket.Sprint{
				DomainEntity:    domainlayer.DomainEntity{Id: iterIdGen.Generate(data.Options.ConnectionId, iter.Id)},
				Url:             fmt.Sprintf("https://www.tapd.cn/%d/prong/iterations/view/%d", iter.WorkspaceId, iter.Id),
				Status:          strings.ToUpper(iter.Status),
				Name:            iter.Name,
				StartedDate:     (*time.Time)(iter.Startdate),
				EndedDate:       (*time.Time)(iter.Enddate),
				OriginalBoardID: workspaceIdGen.Generate(iter.ConnectionId, iter.WorkspaceId),
				CompletedDate:   (*time.Time)(iter.Completed),
			}
			results := make([]interface{}, 0)
			results = append(results, domainIter)
			boardSprint := &ticket.BoardSprint{
				BoardId:  domainIter.OriginalBoardID,
				SprintId: domainIter.Id,
			}
			results = append(results, boardSprint)
			return results, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}

var ConvertIterationMeta = core.SubTaskMeta{
	Name:             "convertIteration",
	EntryPoint:       ConvertIteration,
	EnabledByDefault: true,
	Description:      "convert Tapd iteration",
}
