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
package chats

import (
	"github.com/arangodb/go-driver"
	"github.com/slntopp/nocloud/pkg/chats/proto"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type ChatsServer struct {
	proto.UnimplementedChatServiceServer
}

func NewChatsServer(log *zap.Logger, db driver.Database, rbmq *amqp.Connection) *ChatsServer {
	return nil
}
