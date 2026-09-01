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
package registry

import (
	"strconv"
	"time"

	"github.com/slntopp/nocloud-proto/access"
	accountspb "github.com/slntopp/nocloud-proto/registry/accounts"
	"github.com/slntopp/nocloud/pkg/graph"
	sc "github.com/slntopp/nocloud/pkg/settings/client"
)

func MakeAccountMessage(acc graph.Account) *accountspb.Account {
	return &accountspb.Account{Uuid: acc.Key, Title: acc.Title}
}

func MergeMaps[K comparable](old map[K]interface{}, new map[K]interface{}) map[K]interface{} {
	result := make(map[K]interface{})
	for key, ov := range old {
		result[key] = ov
	}

	for key, nv := range new {
		if nv == nil {
			delete(result, key)
			continue
		}

		result[key] = nv
	}

	return result
}

// normalizeDateCreate ensures data.date_create is a unix timestamp number.
// Historical WHMCS sync stored it as a string, which breaks AQL numeric filters/sorts.
// If setIfMissing is true and the field is absent/empty, it is set to now (Create/SignUp).
func normalizeDateCreate(m map[string]any, setIfMissing bool) {
	if m == nil {
		return
	}
	v, ok := m["date_create"]
	if !ok || v == nil || v == "" {
		if setIfMissing {
			m["date_create"] = float64(time.Now().Unix())
		}
		return
	}
	switch t := v.(type) {
	case float64:
		if t == 0 && setIfMissing {
			m["date_create"] = float64(time.Now().Unix())
		}
	case string:
		n, err := strconv.ParseFloat(t, 64)
		if err != nil || n == 0 {
			if setIfMissing {
				m["date_create"] = float64(time.Now().Unix())
			} else {
				delete(m, "date_create")
			}
		} else {
			m["date_create"] = n
		}
	case int64:
		m["date_create"] = float64(t)
	case int:
		m["date_create"] = float64(t)
	case int32:
		m["date_create"] = float64(t)
	default:
		if setIfMissing {
			m["date_create"] = float64(time.Now().Unix())
		}
	}
}

const accountPostCreateSettingsKey = "post-create-account"

type AccountPostCreateSettings struct {
	CreateNamespace bool `json:"create-ns"`
}

const signupKey = "signup"

type SignUpSettings struct {
	Namespace      string   `json:"namespace"`
	AllowedTypes   []string `json:"allowed_types"`
	Enabled        bool     `json:"enabled"`
	EnabledAccount bool     `json:"enabled_account"`
	BaseTaxRate    float64  `json:"base_tax_rate"`
}

var defaultSettings = &sc.Setting[AccountPostCreateSettings]{
	Value:       AccountPostCreateSettings{CreateNamespace: true},
	Description: "Post Account Creation Actions",
	Level:       access.Level_ADMIN,
}

var standartSettings = &sc.Setting[SignUpSettings]{
	Value: SignUpSettings{
		Namespace:      "",
		AllowedTypes:   []string{},
		Enabled:        false,
		EnabledAccount: false,
		BaseTaxRate:    0,
	},
	Description: "Signup Settings",
	Level:       access.Level_ADMIN,
}
