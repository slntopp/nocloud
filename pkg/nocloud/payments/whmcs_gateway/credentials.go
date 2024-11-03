package whmcs_gateway

import (
	"context"
	"encoding/json"
	redisdb "github.com/slntopp/nocloud/pkg/nocloud/redis"
)

const whmcsRedisKey = "_settings:whmcs"

type WhmcsData struct {
	WhmcsUser     string `json:"user"`
	WhmcsPassHash string `json:"pass_hash"`
	WhmcsBaseUrl  string `json:"api"`
	DangerMode    bool   `json:"danger_mode"`
	TrustedIP     string `json:"trusted_ip"`
}

func GetWhmcsCredentials(rdb redisdb.Client) (WhmcsData, error) {
	var whmcsData WhmcsData
	if keys, err := rdb.HGetAll(context.Background(), whmcsRedisKey).Result(); err == nil {
		if err := json.Unmarshal([]byte(keys["value"]), &whmcsData); err != nil {
			return whmcsData, err
		}
	} else {
		return whmcsData, err
	}

	return whmcsData, nil
}
