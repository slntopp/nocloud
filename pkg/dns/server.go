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
package dns

import (
	redis "github.com/go-redis/redis/v8"
	pb "github.com/slntopp/nocloud/pkg/dns/proto"
	"go.uber.org/zap"
)

const KEYS_PREFIX = "_dns"

type DNSServer struct {
	pb.UnimplementedDNSServer

	log *zap.Logger
	rdb *redis.Client
}

func NewDNSServer(log *zap.Logger, rdb *redis.Client) *DNSServer {
	return &DNSServer{
		log: log.Named("DNSServer"), rdb: rdb,
	}
}

