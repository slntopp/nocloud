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
package nocloud

import (
	"context"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const NOCLOUD_ACCOUNT_CLAIM = "account"

type ContextKey string;
const NoCloudAccount = ContextKey("account");

func NewLogger() (log *zap.Logger) {
	viper.SetDefault("LOG_LEVEL", 0)
	level := viper.GetInt("LOG_LEVEL")

	atom := zap.NewAtomicLevel()
	atom.SetLevel(zapcore.Level(level))

	encoderCfg := zap.NewProductionEncoderConfig()
	return zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))
}

func ValidateMetadata(ctx context.Context, log *zap.Logger) (context.Context, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Error("Failed to get metadata from context")
		return ctx, status.Error(codes.Aborted, "Failed to get metadata from context")
	}

	//Check for Authentication
	requestor := md.Get(NOCLOUD_ACCOUNT_CLAIM)
	if requestor == nil {
		log.Error("Failed to authenticate account")
		return ctx, status.Error(codes.Unauthenticated, "Not authenticated")
	}
	ctx = context.WithValue(ctx, NoCloudAccount, requestor[0])

	return ctx, nil
}