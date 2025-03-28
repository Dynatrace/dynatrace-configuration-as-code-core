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

// checks if the SubscriptionListDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SubscriptionListDto{}

// SubscriptionListDto struct for SubscriptionListDto
type SubscriptionListDto struct {
	// A list of subscriptions of the account.
	Data                 []SubscriptionSummaryDto `json:"data"`
	AdditionalProperties map[string]interface{}
}

type _SubscriptionListDto SubscriptionListDto

// NewSubscriptionListDto instantiates a new SubscriptionListDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSubscriptionListDto(data []SubscriptionSummaryDto) *SubscriptionListDto {
	this := SubscriptionListDto{}
	this.Data = data
	return &this
}

// NewSubscriptionListDtoWithDefaults instantiates a new SubscriptionListDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSubscriptionListDtoWithDefaults() *SubscriptionListDto {
	this := SubscriptionListDto{}
	return &this
}

// GetData returns the Data field value
func (o *SubscriptionListDto) GetData() []SubscriptionSummaryDto {
	if o == nil {
		var ret []SubscriptionSummaryDto
		return ret
	}

	return o.Data
}

// GetDataOk returns a tuple with the Data field value
// and a boolean to check if the value has been set.
func (o *SubscriptionListDto) GetDataOk() ([]SubscriptionSummaryDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Data, true
}

// SetData sets field value
func (o *SubscriptionListDto) SetData(v []SubscriptionSummaryDto) {
	o.Data = v
}

func (o SubscriptionListDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SubscriptionListDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["data"] = o.Data

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *SubscriptionListDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"data",
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

	varSubscriptionListDto := _SubscriptionListDto{}

	err = json.Unmarshal(data, &varSubscriptionListDto)

	if err != nil {
		return err
	}

	*o = SubscriptionListDto(varSubscriptionListDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "data")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableSubscriptionListDto struct {
	value *SubscriptionListDto
	isSet bool
}

func (v NullableSubscriptionListDto) Get() *SubscriptionListDto {
	return v.value
}

func (v *NullableSubscriptionListDto) Set(val *SubscriptionListDto) {
	v.value = val
	v.isSet = true
}

func (v NullableSubscriptionListDto) IsSet() bool {
	return v.isSet
}

func (v *NullableSubscriptionListDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSubscriptionListDto(val *SubscriptionListDto) *NullableSubscriptionListDto {
	return &NullableSubscriptionListDto{value: val, isSet: true}
}

func (v NullableSubscriptionListDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSubscriptionListDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
