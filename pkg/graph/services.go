/*
Copyright © 2021 Nikita Ivanovski info@slnt-opp.xyz

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package graph

import (
	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
)

const (
	SERVICES_COL = "Services"
	NS2SERV = NAMESPACES_COL + "2" + SERVICES_COL
)

type Service struct {
	Title string `json:"title"`
	Config interface{} `json:"config"`
	Instances []interface{} `json:"instances"`

	driver.DocumentMeta
}

type ServicesController struct {
	col driver.Collection // Services Collection

	log *zap.Logger
}

func NewServicesController(log *zap.Logger, col driver.Collection) ServicesController {
	return ServicesController{log: log, col: col}
}
