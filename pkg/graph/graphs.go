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