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
class NamespacesServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Nocloud\Registry\Namespaces\CreateRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Create(\Nocloud\Registry\Namespaces\CreateRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.NamespacesService/Create',
        $argument,
        ['\Nocloud\Registry\Namespaces\CreateResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Namespaces\ListRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function List(\Nocloud\Registry\Namespaces\ListRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.NamespacesService/List',
        $argument,
        ['\Nocloud\Registry\Namespaces\ListResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Namespaces\JoinRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Join(\Nocloud\Registry\Namespaces\JoinRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.NamespacesService/Join',
        $argument,
        ['\Nocloud\Registry\Namespaces\JoinResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Namespaces\LinkRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Link(\Nocloud\Registry\Namespaces\LinkRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.NamespacesService/Link',
        $argument,
        ['\Nocloud\Registry\Namespaces\LinkResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Registry\Namespaces\DeleteRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Delete(\Nocloud\Registry\Namespaces\DeleteRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.registry.NamespacesService/Delete',
        $argument,
        ['\Nocloud\Registry\Namespaces\DeleteResponse', 'decode'],
        $metadata, $options);
    }

}
