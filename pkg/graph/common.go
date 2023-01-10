/*
Copyright © 2021-2023 Nikita Ivanovski info@slnt-opp.xyz

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
	"context"

	"github.com/arangodb/go-driver"
	"go.uber.org/zap"
)

func GraphGetEnsure(log *zap.Logger, ctx context.Context, db driver.Database, name string) driver.Graph {
	exists, err := db.GraphExists(ctx, name)
	if err != nil {
		log.Fatal("Error checking graph existence", zap.Error(err))
	}
	if !exists {
		log.Info("Creating permissions graph")
		graph, err := db.CreateGraphV2(ctx, name, nil)
		if err != nil {
			log.Fatal("Error creating permissions graph", zap.Error(err))
		}
		return graph
	}

	graph, err := db.Graph(ctx, name)
	if err != nil {
		log.Fatal("Error getting graph", zap.Error(err))
	}
	return graph
}

func GraphGetVertexEnsure(log *zap.Logger, ctx context.Context, db driver.Database, graph driver.Graph, name string) (col driver.Collection) {
	exists, err := db.CollectionExists(ctx, name)
	if err != nil {
		log.Fatal("Error checking if collection exists", zap.Error(err))
	}
	if !exists {
		col, err = db.CreateCollection(ctx, name, &driver.CreateCollectionOptions{
			KeyOptions: &driver.CollectionKeyOptions{AllowUserKeys: true, Type: "uuid"},
		})
		if err != nil {
			log.Fatal("Error creating collection", zap.Error(err))
		}
		return col
	}
	col, err = db.Collection(ctx, name)
	if err != nil {
		log.Fatal("Error getting collection", zap.Error(err))
	}
	return col
}

func GraphGetEdgeEnsure(log *zap.Logger, ctx context.Context, graph driver.Graph, name, from, to string) driver.Collection {
	exists, err := graph.EdgeCollectionExists(ctx, name)
	if err != nil {
		log.Fatal("Error checking if edge collection exists", zap.Error(err))
	}
	if !exists {
		col, err := graph.CreateEdgeCollection(ctx, name, driver.VertexConstraints{
			From: []string{from}, To: []string{to},
		})
		if err != nil {
			log.Fatal("Error creating collection", zap.Error(err))
		}
		return col
	}

	err = graph.SetVertexConstraints(ctx, name, driver.VertexConstraints{
		From: []string{from}, To: []string{to},
	})
	if err != nil {
		log.Fatal("Error setting vertex constraints for edge collection",
			zap.String("col", name), zap.String("from", from),
			zap.String("to", to), zap.Error(err),
		)
	}

	col, _, err := graph.EdgeCollection(ctx, name)
	if err != nil {
		log.Fatal("Error getting collection", zap.Error(err))
	}
	return col
}

func GetEnsureCollection(log *zap.Logger, ctx context.Context, db driver.Database, name string) driver.Collection {
	exists, err := db.CollectionExists(ctx, name)
	if err != nil {
		log.Fatal("Error checking if collection exists", zap.Error(err))
	}
	if !exists {
		col, err := db.CreateCollection(ctx, name, &driver.CreateCollectionOptions{
			KeyOptions: &driver.CollectionKeyOptions{AllowUserKeys: true, Type: "uuid"},
		})
		if err != nil {
			log.Fatal("Error creating collection", zap.Error(err))
		}
		return col
	}

	col, err := db.Collection(ctx, name)
	if err != nil {
		log.Fatal("Error getting collection", zap.Error(err))
	}
	return col
}
