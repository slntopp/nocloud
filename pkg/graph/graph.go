package graph

import (
	"context"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"go.uber.org/zap"
)


var (
	DB_NAME = "nocloud"
)

func CheckAndRegisterCollections(log *zap.Logger, db driver.Database, collections []string) {
	for _, col := range collections {
		log.Debug("Checking Collection existence", zap.String("collection", col))
		exists, err := db.CollectionExists(nil, col)
		if err != nil {
			log.Fatal("Failed to check collection", zap.Any(col, err))
		}
		log.Debug("Collection " + col, zap.Bool("Exists", exists))
		if !exists {
			log.Debug("Creating", zap.String("collection", col))
			_, err := db.CreateCollection(nil, col, nil)
			if err != nil {
				log.Fatal("Failed to create collection", zap.Any(col, err))
			}
		}
	}
}

func CheckAndRegisterGraph(log *zap.Logger, db driver.Database, graph NoCloudGraphSchema) {
	graphExists, err := db.GraphExists(nil, graph.Name)
	if err != nil {
		log.Fatal("Failed to check graph", zap.Any(graph.Name, err))
	}
	log.Debug("Graph Permissions", zap.Bool("Exists", graphExists))

	if graphExists {
		return
	}
	log.Debug("Creating", zap.String("graph", graph.Name))
	edges := make([]driver.EdgeDefinition, 0)
	for _, edge := range graph.Edges {
		edges = append(edges, driver.EdgeDefinition{
			Collection: strings.Join(edge, "2"),
			From: []string{edge[0]}, To: []string{edge[1]},
		})
	}

	var options driver.CreateGraphOptions
	options.EdgeDefinitions = edges

	_, err = db.CreateGraph(nil, graph.Name, &options)
	if err != nil {
		log.Fatal("Failed to create Graph", zap.Any(graph.Name, err))
	}
}

func InitDB(log *zap.Logger, dbHost, dbCred, rootPass string) {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://" + dbCred + "@" + dbHost},
	})
	if err != nil {
		log.Fatal("Error creating connection to DB", zap.Error(err))
	}

	c, err := driver.NewClient(driver.ClientConfig{
		Connection: conn,
	})
	if err != nil {
		log.Fatal("Error creating driver instance for DB", zap.Error(err))
	}

	// Checking if DB exists and creating it if not
	log.Debug("Checking if DB exists")
	dbExists, err := c.DatabaseExists(nil, DB_NAME)
	if err != nil {
		log.Fatal("Error checking if DataBase exists", zap.Error(err))
	}
	log.Debug("DataBase", zap.Bool("Exists", dbExists))
	
	var db driver.Database
	if !dbExists {
		db, err = c.CreateDatabase(nil, DB_NAME, nil)
		if err != nil {
			log.Fatal("Error creating DataBase", zap.Error(err))
		}
	}
	db, err = c.Database(nil, DB_NAME)

	CheckAndRegisterCollections(log, db, COLLECTIONS)

	for _, graph := range GRAPHS_SCHEMAS {
		CheckAndRegisterGraph(log, db, graph)
	}

	col, _ := db.Collection(nil, ACCOUNTS_COL)
	account_ctrl := AccountsController{
		col: col, log: log,
	}
	rootExists, err := col.DocumentExists(nil, "0")
	if err != nil {
		log.Fatal("Failed to check root Account", zap.Any("error", err))
	}

	if !rootExists {
		meta, err := col.CreateDocument(nil, Account{ 
			Title: "root",
			DocumentMeta: driver.DocumentMeta { Key: "0" },
		})
		if err != nil {
			log.Fatal("Failed to create root Account", zap.Any("error", err))
		}
		log.Debug("Create root Account", zap.Any("result", meta))
	}

	var root Account
	var meta driver.DocumentMeta
	meta, err = col.ReadDocument(nil, "0", &root)
	if err != nil {
		log.Fatal("Failed to get root", zap.Any("error", err))
	}
	root.DocumentMeta = meta
	log.Debug("Got account", zap.Any("result", root))

	ctx := context.WithValue(context.Background(), nocloud.NoCloudAccount, "0")

	col, _ = db.Collection(ctx, NAMESPACES_COL)
	rootNSExists, err := col.DocumentExists(ctx, "0")
	if err != nil {
		log.Fatal("Failed to check root Account", zap.Any("error", err))
	}

	if !rootNSExists {
		meta, err := col.CreateDocument(ctx, Namespace{ 
			Title: "root",
			DocumentMeta: driver.DocumentMeta { Key: "0" },
		})
		if err != nil {
			log.Fatal("Failed to create root Namespace", zap.Any("error", err))
		}
		log.Debug("Create root Account", zap.Any("result", meta))
	}

	var rootNS Namespace
	meta, err = col.ReadDocument(ctx, "0", &rootNS)
	if err != nil {
		log.Fatal("Failed to get root", zap.Any("error", err))
	}
	rootNS.DocumentMeta = meta
	log.Debug("Got namespace", zap.Any("result", rootNS))

	log.Debug("Creating superadmin permission root -> root(ns)")
	col, _ = db.Collection(ctx, ACCOUNTS_COL + "2" + NAMESPACES_COL)
	err = root.LinkNamespace(ctx, col, rootNS, 4)
	if err != nil {
		log.Error("Error while creating edge", zap.Error(err))
	}

	log.Debug("Creating credentials(standard) for root account")
	cred, err := NewStandardCredentials("root", rootPass)
	col, _ = db.Collection(ctx, CREDENTIALS_COL)
	account_ctrl.cred = col
	col, _ = db.Collection(ctx,  ACCOUNTS_COL + "2" + CREDENTIALS_COL)
	err = account_ctrl.SetCredentials(ctx, root, col, cred)
	if err != nil {
		log.Error("Error while creating credentials", zap.Error(err))
	}

	log.Debug("Checking root credentials")
	_, r := account_ctrl.Authorize(ctx, "standard", "root", rootPass)
	log.Debug("Result", zap.Bool("Authorized", r))
}