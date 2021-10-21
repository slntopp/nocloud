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

type NoCloudGraphSchema struct {
	Name string
	Edges [][]string
}

var COLLECTIONS = []string{ACCOUNTS_COL, NAMESPACES_COL, "Services", "Instances", CREDENTIALS_COL}

var PERMISSIONS_GRAPH = NoCloudGraphSchema{
	Name: "Permissions",
	Edges: [][]string{
		{ACCOUNTS_COL, NAMESPACES_COL},
		{NAMESPACES_COL, ACCOUNTS_COL},
		{NAMESPACES_COL, "Services"},
		{"Services", "Instances"},
	},
}
var CREDENTIALS_GRAPH = NoCloudGraphSchema{
	Name: "Credentials",
	Edges: [][]string{
		{ACCOUNTS_COL, CREDENTIALS_COL},
	},
}

var GRAPHS_SCHEMAS = []NoCloudGraphSchema{
	PERMISSIONS_GRAPH, CREDENTIALS_GRAPH,
}