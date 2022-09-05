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
package graph

import (
	"context"

	"github.com/arangodb/go-driver"
	chatpb "github.com/slntopp/nocloud/pkg/chats/proto"
	"github.com/slntopp/nocloud/pkg/nocloud/schema"
	"go.uber.org/zap"
)

type Chat struct {
	*chatpb.Chat
	driver.DocumentMeta
}
type ChatsController struct {
	log   *zap.Logger
	col   driver.Collection
	graph driver.Graph
}

type ChatMessage struct {
	*chatpb.ChatMessage
	driver.DocumentMeta
}

type ChatsMessagesController struct {
	log     *zap.Logger
	col     driver.Collection
	cht2msg driver.Collection
	graph   driver.Graph
}

func NewChatsController(logger *zap.Logger, db driver.Database) ChatsController {
	ctx := context.TODO()
	log := logger.Named("ChatsController")

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.CHATS_COL)

	GraphGetEdgeEnsure(log, ctx, graph, schema.NS2CHTS, schema.NAMESPACES_COL, schema.CHATS_COL)

	return ChatsController{log: log, col: col, graph: graph}
}

func NewChatsMessagesController(logger *zap.Logger, db driver.Database) ChatsMessagesController {
	ctx := context.TODO()
	log := logger.Named("ChatsMessagesController")

	graph := GraphGetEnsure(log, ctx, db, schema.PERMISSIONS_GRAPH.Name)
	col := GraphGetVertexEnsure(log, ctx, db, graph, schema.CHATS_MESSAGES_COL)

	cht2msg := GraphGetEdgeEnsure(log, ctx, graph, schema.CHT2MSG, schema.CHATS_COL, schema.CHATS_MESSAGES_COL)

	return ChatsMessagesController{log: log, col: col, cht2msg: cht2msg, graph: graph}
}

func (ctrl *ChatsController) Get(ctx context.Context, id string) (Chat, error) {
	logger := ctrl.log.Named("GetChat")
	logger.Info("Getting chat", zap.String("id", id))

	var chat chatpb.Chat
	meta, err := ctrl.col.ReadDocument(ctx, id, &chat)
	if err != nil {
		return Chat{}, err
	}
	chat.Uuid = meta.ID.Key()

	return Chat{&chat, meta}, nil
}

func (ctrl *ChatsMessagesController) Get(ctx context.Context, id string) (ChatMessage, error) {
	logger := ctrl.log.Named("GetChatMessage")
	logger.Info("Getting chat message", zap.String("id", id))

	var msg chatpb.ChatMessage
	meta, err := ctrl.col.ReadDocument(ctx, id, &msg)
	if err != nil {
		return ChatMessage{}, err
	}
	msg.Uuid = meta.ID.Key()

	return ChatMessage{&msg, meta}, nil
}
