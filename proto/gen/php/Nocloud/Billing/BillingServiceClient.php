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
class BillingServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Nocloud\Billing\Plan $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function CreatePlan(\Nocloud\Billing\Plan $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.BillingService/CreatePlan',
        $argument,
        ['\Nocloud\Billing\Plan', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\Plan $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function UpdatePlan(\Nocloud\Billing\Plan $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.BillingService/UpdatePlan',
        $argument,
        ['\Nocloud\Billing\Plan', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\Plan $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function GetPlan(\Nocloud\Billing\Plan $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.BillingService/GetPlan',
        $argument,
        ['\Nocloud\Billing\Plan', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\ListRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function ListPlans(\Nocloud\Billing\ListRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.BillingService/ListPlans',
        $argument,
        ['\Nocloud\Billing\ListResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\Plan $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function DeletePlan(\Nocloud\Billing\Plan $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.BillingService/DeletePlan',
        $argument,
        ['\Nocloud\Billing\Plan', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\Transaction $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function CreateTransaction(\Nocloud\Billing\Transaction $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.BillingService/CreateTransaction',
        $argument,
        ['\Nocloud\Billing\Transaction', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\GetTransactionsRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function GetTransactions(\Nocloud\Billing\GetTransactionsRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.BillingService/GetTransactions',
        $argument,
        ['\Nocloud\Billing\Transactions', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\Transaction $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function GetRecords(\Nocloud\Billing\Transaction $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.BillingService/GetRecords',
        $argument,
        ['\Nocloud\Billing\Records', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Billing\ReprocessTransactionsRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Reprocess(\Nocloud\Billing\ReprocessTransactionsRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.billing.BillingService/Reprocess',
        $argument,
        ['\Nocloud\Billing\Transactions', 'decode'],
        $metadata, $options);
    }

}
