package graph

type NoCloudGraphSchema struct {
	Name string
	Collections []string
	Edges [][]string
}

var COLLECTIONS = []string{ACCOUNTS_COL, NAMESPACES_COL, "Services", "Instances"}

var PERMISSIONS_GRAPH = NoCloudGraphSchema{
	Name: "Permissions",
	Edges: [][]string{
		{ACCOUNTS_COL, NAMESPACES_COL},
		{NAMESPACES_COL, ACCOUNTS_COL},
		{NAMESPACES_COL, "Services"},
		{"Services", "Instances"},
	},
}

var GRAPHS_SCHEMAS = []NoCloudGraphSchema{
	PERMISSIONS_GRAPH,
}