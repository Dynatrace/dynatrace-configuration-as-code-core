/*
Dynatrace Account Management API

The enterprise management API for Dynatrace SaaS enables automation of operational tasks related to user access and environment lifecycle management.

API version: 1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package accountmanagement

import (
	"encoding/json"
	"fmt"
)

// checks if the SubscriptionEnvironmentUsageV3Dto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SubscriptionEnvironmentUsageV3Dto{}

// SubscriptionEnvironmentUsageV3Dto struct for SubscriptionEnvironmentUsageV3Dto
type SubscriptionEnvironmentUsageV3Dto struct {
	// The UUID of the Managed cluster.
	ClusterId string `json:"clusterId"`
	// The UUID of the environment.
	EnvironmentId string `json:"environmentId"`
	// Subscription usage information for the environment.
	Usage                []SubscriptionUsageDto `json:"usage"`
	AdditionalProperties map[string]interface{}
}

type _SubscriptionEnvironmentUsageV3Dto SubscriptionEnvironmentUsageV3Dto

// NewSubscriptionEnvironmentUsageV3Dto instantiates a new SubscriptionEnvironmentUsageV3Dto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSubscriptionEnvironmentUsageV3Dto(clusterId string, environmentId string, usage []SubscriptionUsageDto) *SubscriptionEnvironmentUsageV3Dto {
	this := SubscriptionEnvironmentUsageV3Dto{}
	this.ClusterId = clusterId
	this.EnvironmentId = environmentId
	this.Usage = usage
	return &this
}

// NewSubscriptionEnvironmentUsageV3DtoWithDefaults instantiates a new SubscriptionEnvironmentUsageV3Dto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSubscriptionEnvironmentUsageV3DtoWithDefaults() *SubscriptionEnvironmentUsageV3Dto {
	this := SubscriptionEnvironmentUsageV3Dto{}
	return &this
}

// GetClusterId returns the ClusterId field value
func (o *SubscriptionEnvironmentUsageV3Dto) GetClusterId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ClusterId
}

// GetClusterIdOk returns a tuple with the ClusterId field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentUsageV3Dto) GetClusterIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ClusterId, true
}

// SetClusterId sets field value
func (o *SubscriptionEnvironmentUsageV3Dto) SetClusterId(v string) {
	o.ClusterId = v
}

// GetEnvironmentId returns the EnvironmentId field value
func (o *SubscriptionEnvironmentUsageV3Dto) GetEnvironmentId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.EnvironmentId
}

// GetEnvironmentIdOk returns a tuple with the EnvironmentId field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentUsageV3Dto) GetEnvironmentIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EnvironmentId, true
}

// SetEnvironmentId sets field value
func (o *SubscriptionEnvironmentUsageV3Dto) SetEnvironmentId(v string) {
	o.EnvironmentId = v
}

// GetUsage returns the Usage field value
func (o *SubscriptionEnvironmentUsageV3Dto) GetUsage() []SubscriptionUsageDto {
	if o == nil {
		var ret []SubscriptionUsageDto
		return ret
	}

	return o.Usage
}

// GetUsageOk returns a tuple with the Usage field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentUsageV3Dto) GetUsageOk() ([]SubscriptionUsageDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Usage, true
}

// SetUsage sets field value
func (o *SubscriptionEnvironmentUsageV3Dto) SetUsage(v []SubscriptionUsageDto) {
	o.Usage = v
}

func (o SubscriptionEnvironmentUsageV3Dto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SubscriptionEnvironmentUsageV3Dto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["clusterId"] = o.ClusterId
	toSerialize["environmentId"] = o.EnvironmentId
	toSerialize["usage"] = o.Usage

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *SubscriptionEnvironmentUsageV3Dto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"clusterId",
		"environmentId",
		"usage",
	}

	allProperties := make(map[string]interface{})

	err = json.Unmarshal(data, &allProperties)

	if err != nil {
		return err
	}

	for _, requiredProperty := range requiredProperties {
		if _, exists := allProperties[requiredProperty]; !exists {
			return fmt.Errorf("no value given for required property %v", requiredProperty)
		}
	}

	varSubscriptionEnvironmentUsageV3Dto := _SubscriptionEnvironmentUsageV3Dto{}

	err = json.Unmarshal(data, &varSubscriptionEnvironmentUsageV3Dto)

	if err != nil {
		return err
	}

	*o = SubscriptionEnvironmentUsageV3Dto(varSubscriptionEnvironmentUsageV3Dto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "clusterId")
		delete(additionalProperties, "environmentId")
		delete(additionalProperties, "usage")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableSubscriptionEnvironmentUsageV3Dto struct {
	value *SubscriptionEnvironmentUsageV3Dto
	isSet bool
}

func (v NullableSubscriptionEnvironmentUsageV3Dto) Get() *SubscriptionEnvironmentUsageV3Dto {
	return v.value
}

func (v *NullableSubscriptionEnvironmentUsageV3Dto) Set(val *SubscriptionEnvironmentUsageV3Dto) {
	v.value = val
	v.isSet = true
}

func (v NullableSubscriptionEnvironmentUsageV3Dto) IsSet() bool {
	return v.isSet
}

func (v *NullableSubscriptionEnvironmentUsageV3Dto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSubscriptionEnvironmentUsageV3Dto(val *SubscriptionEnvironmentUsageV3Dto) *NullableSubscriptionEnvironmentUsageV3Dto {
	return &NullableSubscriptionEnvironmentUsageV3Dto{value: val, isSet: true}
}

func (v NullableSubscriptionEnvironmentUsageV3Dto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSubscriptionEnvironmentUsageV3Dto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
