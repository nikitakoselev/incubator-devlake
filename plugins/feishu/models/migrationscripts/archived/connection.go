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

package archived

import (
	commonArchived "github.com/apache/incubator-devlake/models/migrationscripts/archived"
)

type FeishuConnection struct {
	commonArchived.Model
	Name             string `gorm:"type:varchar(100);uniqueIndex" json:"name" validate:"required"`
	Endpoint         string `mapstructure:"endpoint" env:"GITHUB_ENDPOINT" validate:"required"`
	Proxy            string `mapstructure:"proxy" env:"GITHUB_PROXY"`
	RateLimitPerHour int    `comment:"api request rate limit per hour"`
	AppId            string `mapstructure:"app_id" validate:"required" json:"app_id"`
	SecretKey        string `mapstructure:"secret_key" validate:"required" json:"secret_key" encrypt:"yes"`
}

func (FeishuConnection) TableName() string {
	return "_tool_feishu_connections"
}
