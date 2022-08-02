<?php
// GENERATED CODE -- DO NOT EDIT!

// Original file comments:
//
// Copyright © 2021-2022 Nikita Ivanovski info@slnt-opp.xyz
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
namespace Nocloud\Settings;

/**
 */
class SettingsServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Nocloud\Settings\GetRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Get(\Nocloud\Settings\GetRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.settings.SettingsService/Get',
        $argument,
        ['\Google\Protobuf\Struct', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Settings\PutRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Put(\Nocloud\Settings\PutRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.settings.SettingsService/Put',
        $argument,
        ['\Nocloud\Settings\PutResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * rpc Sub(nocloud.settings.SubRequest) returns (stream nocloud.settings.SubRequest);
     * @param \Nocloud\Settings\KeysRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Keys(\Nocloud\Settings\KeysRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.settings.SettingsService/Keys',
        $argument,
        ['\Nocloud\Settings\KeysResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Settings\DeleteRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Delete(\Nocloud\Settings\DeleteRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.settings.SettingsService/Delete',
        $argument,
        ['\Nocloud\Settings\DeleteResponse', 'decode'],
        $metadata, $options);
    }

}
