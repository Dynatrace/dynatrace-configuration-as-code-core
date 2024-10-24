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

// checks if the SubscriptionEnvironmentCostV2Dto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SubscriptionEnvironmentCostV2Dto{}

// SubscriptionEnvironmentCostV2Dto struct for SubscriptionEnvironmentCostV2Dto
type SubscriptionEnvironmentCostV2Dto struct {
	// The UUID of the environment
	EnvironmentId string `json:"environmentId"`
	// A list of subscription cost for the environment.
	Cost                 []SubscriptionCapabilityCostDto `json:"cost"`
	AdditionalProperties map[string]interface{}
}

type _SubscriptionEnvironmentCostV2Dto SubscriptionEnvironmentCostV2Dto

// NewSubscriptionEnvironmentCostV2Dto instantiates a new SubscriptionEnvironmentCostV2Dto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSubscriptionEnvironmentCostV2Dto(environmentId string, cost []SubscriptionCapabilityCostDto) *SubscriptionEnvironmentCostV2Dto {
	this := SubscriptionEnvironmentCostV2Dto{}
	this.EnvironmentId = environmentId
	this.Cost = cost
	return &this
}

// NewSubscriptionEnvironmentCostV2DtoWithDefaults instantiates a new SubscriptionEnvironmentCostV2Dto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSubscriptionEnvironmentCostV2DtoWithDefaults() *SubscriptionEnvironmentCostV2Dto {
	this := SubscriptionEnvironmentCostV2Dto{}
	return &this
}

// GetEnvironmentId returns the EnvironmentId field value
func (o *SubscriptionEnvironmentCostV2Dto) GetEnvironmentId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.EnvironmentId
}

// GetEnvironmentIdOk returns a tuple with the EnvironmentId field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentCostV2Dto) GetEnvironmentIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EnvironmentId, true
}

// SetEnvironmentId sets field value
func (o *SubscriptionEnvironmentCostV2Dto) SetEnvironmentId(v string) {
	o.EnvironmentId = v
}

// GetCost returns the Cost field value
func (o *SubscriptionEnvironmentCostV2Dto) GetCost() []SubscriptionCapabilityCostDto {
	if o == nil {
		var ret []SubscriptionCapabilityCostDto
		return ret
	}

	return o.Cost
}

// GetCostOk returns a tuple with the Cost field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentCostV2Dto) GetCostOk() ([]SubscriptionCapabilityCostDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Cost, true
}

// SetCost sets field value
func (o *SubscriptionEnvironmentCostV2Dto) SetCost(v []SubscriptionCapabilityCostDto) {
	o.Cost = v
}

func (o SubscriptionEnvironmentCostV2Dto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SubscriptionEnvironmentCostV2Dto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["environmentId"] = o.EnvironmentId
	toSerialize["cost"] = o.Cost

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *SubscriptionEnvironmentCostV2Dto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"environmentId",
		"cost",
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

	varSubscriptionEnvironmentCostV2Dto := _SubscriptionEnvironmentCostV2Dto{}

	err = json.Unmarshal(data, &varSubscriptionEnvironmentCostV2Dto)

	if err != nil {
		return err
	}

	*o = SubscriptionEnvironmentCostV2Dto(varSubscriptionEnvironmentCostV2Dto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "environmentId")
		delete(additionalProperties, "cost")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableSubscriptionEnvironmentCostV2Dto struct {
	value *SubscriptionEnvironmentCostV2Dto
	isSet bool
}

func (v NullableSubscriptionEnvironmentCostV2Dto) Get() *SubscriptionEnvironmentCostV2Dto {
	return v.value
}

func (v *NullableSubscriptionEnvironmentCostV2Dto) Set(val *SubscriptionEnvironmentCostV2Dto) {
	v.value = val
	v.isSet = true
}

func (v NullableSubscriptionEnvironmentCostV2Dto) IsSet() bool {
	return v.isSet
}

func (v *NullableSubscriptionEnvironmentCostV2Dto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSubscriptionEnvironmentCostV2Dto(val *SubscriptionEnvironmentCostV2Dto) *NullableSubscriptionEnvironmentCostV2Dto {
	return &NullableSubscriptionEnvironmentCostV2Dto{value: val, isSet: true}
}

func (v NullableSubscriptionEnvironmentCostV2Dto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSubscriptionEnvironmentCostV2Dto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
