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
namespace Nocloud\Billing;

/**
 */
class RecordsServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Nocloud\Billing\GetActiveRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function GetActive(\Nocloud\Billing\GetActiveRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.RecordsService/GetActive',
        $argument,
        ['\Nocloud\Billing\Records', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\Records $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Create(\Nocloud\Billing\Records $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.RecordsService/Create',
        $argument,
        ['\Nocloud\Billing\Records', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\Records $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Update(\Nocloud\Billing\Records $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.RecordsService/Update',
        $argument,
        ['\Nocloud\Billing\Records', 'decode'],
        $metadata, $options);
    }

}
