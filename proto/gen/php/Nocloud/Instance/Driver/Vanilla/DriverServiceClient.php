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
namespace Nocloud\Instance\Driver\Vanilla;

/**
 */
class DriverServiceClient extends \Grpc\BaseStub {

    /**
     * @param string $hostname hostname
     * @param array $opts channel options
     * @param \Grpc\Channel $channel (optional) re-use channel object
     */
    public function __construct($hostname, $opts, $channel = null) {
        parent::__construct($hostname, $opts, $channel);
    }

    /**
     * @param \Nocloud\Instance\Driver\Vanilla\TestServiceProviderConfigRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function TestServiceProviderConfig(\Nocloud\Instance\Driver\Vanilla\TestServiceProviderConfigRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.instance.driver.vanilla.DriverService/TestServiceProviderConfig',
        $argument,
        ['\Nocloud\ServicesProviders\TestResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Instances\TestInstancesGroupConfigRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function TestInstancesGroupConfig(\Nocloud\Instances\TestInstancesGroupConfigRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.instance.driver.vanilla.DriverService/TestInstancesGroupConfig',
        $argument,
        ['\Nocloud\Instances\TestInstancesGroupConfigResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Instance\Driver\Vanilla\GetTypeRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function GetType(\Nocloud\Instance\Driver\Vanilla\GetTypeRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.instance.driver.vanilla.DriverService/GetType',
        $argument,
        ['\Nocloud\Instance\Driver\Vanilla\GetTypeResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Instance\Driver\Vanilla\UpRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Up(\Nocloud\Instance\Driver\Vanilla\UpRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.instance.driver.vanilla.DriverService/Up',
        $argument,
        ['\Nocloud\Instance\Driver\Vanilla\UpResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Instance\Driver\Vanilla\DownRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Down(\Nocloud\Instance\Driver\Vanilla\DownRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.instance.driver.vanilla.DriverService/Down',
        $argument,
        ['\Nocloud\Instance\Driver\Vanilla\DownResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Instance\Driver\Vanilla\MonitoringRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Monitoring(\Nocloud\Instance\Driver\Vanilla\MonitoringRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.instance.driver.vanilla.DriverService/Monitoring',
        $argument,
        ['\Nocloud\Instance\Driver\Vanilla\MonitoringResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Instance\Driver\Vanilla\MonitoringRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function SuspendMonitoring(\Nocloud\Instance\Driver\Vanilla\MonitoringRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.instance.driver.vanilla.DriverService/SuspendMonitoring',
        $argument,
        ['\Nocloud\Instance\Driver\Vanilla\MonitoringResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Instance\Driver\Vanilla\InvokeRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function Invoke(\Nocloud\Instance\Driver\Vanilla\InvokeRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.instance.driver.vanilla.DriverService/Invoke',
        $argument,
        ['\Nocloud\Instances\InvokeResponse', 'decode'],
        $metadata, $options);
    }

    /**
     * @param \Nocloud\Instance\Driver\Vanilla\SpInvokeRequest $argument input argument
     * @param array $metadata metadata
     * @param array $options call options
     * @return \Grpc\UnaryCall
     */
    public function SpInvoke(\Nocloud\Instance\Driver\Vanilla\SpInvokeRequest $argument,
      $metadata = [], $options = []) {
        return $this->_simpleRequest('/nocloud.instance.driver.vanilla.DriverService/SpInvoke',
        $argument,
        ['\Nocloud\ServicesProviders\InvokeResponse', 'decode'],
        $metadata, $options);
    }

}
