package main

import (
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/arangodb/go-driver/http"
	inflog "github.com/infinimesh/infinimesh/pkg/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	log *zap.Logger

	dbHost string
	dbPort string
	dbUser string
	dbPass string

	rootPass string
)

var (
	DB_NAME = "nocloud"
	COLLECTIONS = []string{"Accounts", "Namespaces", "Services", "Instances"}
	EDGES = [][]string{
		{"Accounts", "Namespaces"},
		{"Namespaces", "Accounts"},
		{"Namespaces", "Services"},
		{"Services", "Instances"},
	}
)

func init() {
	logger, err := inflog.NewProdOrDev()
	if err != nil {
		panic(err)
	}
	log = logger

	viper.AutomaticEnv()
	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "8529")
	viper.SetDefault("DB_USER", "root")
	viper.SetDefault("DB_PASS", "openSesame")

	viper.SetDefault("ROOT_PASS", "root")

	dbHost = viper.GetString("DB_HOST")
	dbPort = viper.GetString("DB_PORT")
	dbUser = viper.GetString("DB_USER")
	dbPass = viper.GetString("DB_PASS")

	rootPass = viper.GetString("ROOT_PASS")
}

func main() {
	conn, err := http.NewConnection(http.ConnectionConfig{
		Endpoints: []string{"http://" + dbUser + ":" + dbPass + "@" + dbHost + ":" + dbPort},
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

	for _, col := range COLLECTIONS {
		log.Debug("Checking Collection existence", zap.String("collection", col))
		exists, err := db.CollectionExists(nil, col)
		fmt.Println(exists)
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

	graphExists, err := db.GraphExists(nil, "Permissions")
	if err != nil {
		log.Fatal("Failed to check graph", zap.Any("Permissions", err))
	}
	log.Debug("Graph Permissions", zap.Bool("Exists", graphExists))

	if !graphExists {
		log.Debug("Creating", zap.String("graph", "Permissions"))
		edges := make([]driver.EdgeDefinition, 0)
		for _, edge := range EDGES {
			edges = append(edges, driver.EdgeDefinition{
				Collection: strings.Join(edge, "2"),
				From: []string{edge[0]}, To: []string{edge[1]},
			})
		}

		var options driver.CreateGraphOptions
		options.EdgeDefinitions = edges

		_, err = db.CreateGraph(nil, "Permissions", &options)
		if err != nil {
			log.Fatal("Failed to create Graph", zap.Any("Permissions", err))
		}
	}
}