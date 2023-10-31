/*
Dynatrace Account Management API

The enterprise management API for Dynatrace SaaS enables automation of operational tasks related to user access and environment lifecycle management.

API version: 1.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package accountmanagement

import (
	"encoding/json"
)

// checks if the SubscriptionEnvironmentUsageV2Dto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SubscriptionEnvironmentUsageV2Dto{}

// SubscriptionEnvironmentUsageV2Dto struct for SubscriptionEnvironmentUsageV2Dto
type SubscriptionEnvironmentUsageV2Dto struct {
	// The id of the environment
	EnvironmentId string `json:"environmentId"`
	// A list of subscription usage for the environment.
	Usage []SubscriptionUsageDto `json:"usage"`
}

// NewSubscriptionEnvironmentUsageV2Dto instantiates a new SubscriptionEnvironmentUsageV2Dto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSubscriptionEnvironmentUsageV2Dto(environmentId string, usage []SubscriptionUsageDto) *SubscriptionEnvironmentUsageV2Dto {
	this := SubscriptionEnvironmentUsageV2Dto{}
	this.EnvironmentId = environmentId
	this.Usage = usage
	return &this
}

// NewSubscriptionEnvironmentUsageV2DtoWithDefaults instantiates a new SubscriptionEnvironmentUsageV2Dto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSubscriptionEnvironmentUsageV2DtoWithDefaults() *SubscriptionEnvironmentUsageV2Dto {
	this := SubscriptionEnvironmentUsageV2Dto{}
	return &this
}

// GetEnvironmentId returns the EnvironmentId field value
func (o *SubscriptionEnvironmentUsageV2Dto) GetEnvironmentId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.EnvironmentId
}

// GetEnvironmentIdOk returns a tuple with the EnvironmentId field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentUsageV2Dto) GetEnvironmentIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EnvironmentId, true
}

// SetEnvironmentId sets field value
func (o *SubscriptionEnvironmentUsageV2Dto) SetEnvironmentId(v string) {
	o.EnvironmentId = v
}

// GetUsage returns the Usage field value
func (o *SubscriptionEnvironmentUsageV2Dto) GetUsage() []SubscriptionUsageDto {
	if o == nil {
		var ret []SubscriptionUsageDto
		return ret
	}

	return o.Usage
}

// GetUsageOk returns a tuple with the Usage field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentUsageV2Dto) GetUsageOk() ([]SubscriptionUsageDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Usage, true
}

// SetUsage sets field value
func (o *SubscriptionEnvironmentUsageV2Dto) SetUsage(v []SubscriptionUsageDto) {
	o.Usage = v
}

func (o SubscriptionEnvironmentUsageV2Dto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SubscriptionEnvironmentUsageV2Dto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["environmentId"] = o.EnvironmentId
	toSerialize["usage"] = o.Usage
	return toSerialize, nil
}

type NullableSubscriptionEnvironmentUsageV2Dto struct {
	value *SubscriptionEnvironmentUsageV2Dto
	isSet bool
}

func (v NullableSubscriptionEnvironmentUsageV2Dto) Get() *SubscriptionEnvironmentUsageV2Dto {
	return v.value
}

func (v *NullableSubscriptionEnvironmentUsageV2Dto) Set(val *SubscriptionEnvironmentUsageV2Dto) {
	v.value = val
	v.isSet = true
}

func (v NullableSubscriptionEnvironmentUsageV2Dto) IsSet() bool {
	return v.isSet
}

func (v *NullableSubscriptionEnvironmentUsageV2Dto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSubscriptionEnvironmentUsageV2Dto(val *SubscriptionEnvironmentUsageV2Dto) *NullableSubscriptionEnvironmentUsageV2Dto {
	return &NullableSubscriptionEnvironmentUsageV2Dto{value: val, isSet: true}
}

func (v NullableSubscriptionEnvironmentUsageV2Dto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSubscriptionEnvironmentUsageV2Dto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
