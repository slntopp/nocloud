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
namespace Nocloud\ServicesProviders;

/**
 */
class ServicesProvidersExtentionsServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Nocloud\ServicesProviders\GetTypeRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function GetType(\Nocloud\ServicesProviders\GetTypeRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services_providers.ServicesProvidersExtentionsService/GetType',
        $argument,
        ['\Nocloud\ServicesProviders\GetTypeResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\ServicesProviders\ServicesProvidersExtentionData $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Test(\Nocloud\ServicesProviders\ServicesProvidersExtentionData $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services_providers.ServicesProvidersExtentionsService/Test',
        $argument,
        ['\Nocloud\ServicesProviders\GenericResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\ServicesProviders\ServicesProvidersExtentionData $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Register(\Nocloud\ServicesProviders\ServicesProvidersExtentionData $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services_providers.ServicesProvidersExtentionsService/Register',
        $argument,
        ['\Nocloud\ServicesProviders\GenericResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\ServicesProviders\ServicesProvidersExtentionData $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Update(\Nocloud\ServicesProviders\ServicesProvidersExtentionData $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services_providers.ServicesProvidersExtentionsService/Update',
        $argument,
        ['\Nocloud\ServicesProviders\GenericResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\ServicesProviders\ServicesProvidersExtentionData $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Unregister(\Nocloud\ServicesProviders\ServicesProvidersExtentionData $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.services_providers.ServicesProvidersExtentionsService/Unregister',
        $argument,
        ['\Nocloud\ServicesProviders\GenericResponse', 'decode'],
        $metadata, $options);
    }

}
