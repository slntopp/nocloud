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
namespace Nocloud\Registry;

/**
 */
class AccountsServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Nocloud\Registry\Accounts\TokenRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Token(\Nocloud\Registry\Accounts\TokenRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.AccountsService/Token',
        $argument,
        ['\Nocloud\Registry\Accounts\TokenResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Accounts\SetCredentialsRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function SetCredentials(\Nocloud\Registry\Accounts\SetCredentialsRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.AccountsService/SetCredentials',
        $argument,
        ['\Nocloud\Registry\Accounts\SetCredentialsResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Accounts\CreateRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Create(\Nocloud\Registry\Accounts\CreateRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.AccountsService/Create',
        $argument,
        ['\Nocloud\Registry\Accounts\CreateResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Accounts\Account $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Update(\Nocloud\Registry\Accounts\Account $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.AccountsService/Update',
        $argument,
        ['\Nocloud\Registry\Accounts\UpdateResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Accounts\GetRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Get(\Nocloud\Registry\Accounts\GetRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.AccountsService/Get',
        $argument,
        ['\Nocloud\Registry\Accounts\Account', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Accounts\ListRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function List(\Nocloud\Registry\Accounts\ListRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.AccountsService/List',
        $argument,
        ['\Nocloud\Registry\Accounts\ListResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Accounts\DeleteRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Delete(\Nocloud\Registry\Accounts\DeleteRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.AccountsService/Delete',
        $argument,
        ['\Nocloud\Registry\Accounts\DeleteResponse', 'decode'],
        $metadata, $options);
    }

}
