/*
Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz

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
package schema

const (
	ACCOUNTS_COL = "Accounts"
	ACC2NS = ACCOUNTS_COL + "2" + NAMESPACES_COL
	ACC2CRED = ACCOUNTS_COL + "2" + CREDENTIALS_COL
)

const (
	NAMESPACES_COL = "Namespaces"
	NS2ACC = NAMESPACES_COL + "2" + ACCOUNTS_COL
)

const (
	CREDENTIALS_COL = "Credentials"
	CREDENTIALS_EDGE_COL = ACCOUNTS_COL + "2" + CREDENTIALS_COL
)

const (
	SERVICES_COL = "Services"
	NS2SERV = NAMESPACES_COL + "2" + SERVICES_COL
	SP2SERV = SERVICES_PROVIDERS_COL + "2" + SERVICES_COL
)
const (
	SERVICES_PROVIDERS_COL = "ServicesProviders"
)

type NoCloudGraphSchema struct {
	Name string
	Edges [][]string
}

var COLLECTIONS = []string{
	ACCOUNTS_COL, NAMESPACES_COL, CREDENTIALS_COL,
	SERVICES_PROVIDERS_COL, SERVICES_COL }

var PERMISSIONS_GRAPH = NoCloudGraphSchema{
	Name: "Permissions",
	Edges: [][]string{
		{ACCOUNTS_COL, NAMESPACES_COL},
		{NAMESPACES_COL, ACCOUNTS_COL},
		{NAMESPACES_COL, SERVICES_COL},
	},
}
var CREDENTIALS_GRAPH = NoCloudGraphSchema{
	Name: "Credentials",
	Edges: [][]string{
		{ACCOUNTS_COL, CREDENTIALS_COL},
	},
}
var SERVICES_GRAPH = NoCloudGraphSchema{
	Name: "Services",
	Edges: [][]string{
		{SERVICES_PROVIDERS_COL, SERVICES_COL},
	},
}

var GRAPHS_SCHEMAS = []NoCloudGraphSchema{
	PERMISSIONS_GRAPH, CREDENTIALS_GRAPH, SERVICES_GRAPH,
}