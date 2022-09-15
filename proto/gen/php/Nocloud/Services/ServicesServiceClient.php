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
namespace Nocloud\Services;

/**
 */
class ServicesServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Nocloud\Services\CreateRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function TestConfig(\Nocloud\Services\CreateRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/TestConfig',
        $argument,
        ['\Nocloud\Services\TestConfigResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\CreateRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Create(\Nocloud\Services\CreateRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/Create',
        $argument,
        ['\Nocloud\Services\Service', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\Service $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Update(\Nocloud\Services\Service $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/Update',
        $argument,
        ['\Nocloud\Services\Service', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\DeleteRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Delete(\Nocloud\Services\DeleteRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/Delete',
        $argument,
        ['\Nocloud\Services\DeleteResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\GetRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Get(\Nocloud\Services\GetRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/Get',
        $argument,
        ['\Nocloud\Services\Service', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\ListRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function List(\Nocloud\Services\ListRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/List',
        $argument,
        ['\Nocloud\Services\Services', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\UpRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Up(\Nocloud\Services\UpRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/Up',
        $argument,
        ['\Nocloud\Services\UpResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\DownRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Down(\Nocloud\Services\DownRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/Down',
        $argument,
        ['\Nocloud\Services\DownResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\SuspendRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Suspend(\Nocloud\Services\SuspendRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/Suspend',
        $argument,
        ['\Nocloud\Services\SuspendResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\UnsuspendRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Unsuspend(\Nocloud\Services\UnsuspendRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services.ServicesService/Unsuspend',
        $argument,
        ['\Nocloud\Services\UnsuspendResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Services\StreamRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\ServerStreamingCall
     */
    public function Stream(\Nocloud\Services\StreamRequest $argument,
      $metadata = [], $options = []) {
        return $this->_serverStreamRequest('/nocloud.services.ServicesService/Stream',
        $argument,
        ['\Nocloud\States\ObjectState', 'decode'],
        $metadata, $options);
    }

}
