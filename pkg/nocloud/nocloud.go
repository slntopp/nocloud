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
package nocloud

import (
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const NOCLOUD_ACCOUNT_CLAIM = "account"
const NOCLOUD_ROOT_CLAIM = "root"
const NOCLOUD_SP_CLAIM = "sp"
const NOCLOUD_INSTANCE_CLAIM = "instance"
const NOCLOUD_LOG_LEVEL = zapcore.DebugLevel - 1

type ContextKey string

const NoCloudAccount = ContextKey("account")
const NoCloudRootAccess = ContextKey("root_access")
const NoCloudSp = ContextKey("sp")
const NoCloudInstance = ContextKey("instance")
const NoCloudToken = ContextKey("token")
const TestFromCreate = ContextKey("test_from_create")

func NewLogger() (log *zap.Logger) {
	viper.SetDefault("LOG_LEVEL", NOCLOUD_LOG_LEVEL)
	level := viper.GetInt("LOG_LEVEL")

	atom := zap.NewAtomicLevel()
	atom.SetLevel(zapcore.Level(level))

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	return zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stdout),
		atom,
	))
}
