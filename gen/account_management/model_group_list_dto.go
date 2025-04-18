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

// checks if the GroupListDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GroupListDto{}

// GroupListDto struct for GroupListDto
type GroupListDto struct {
	// The number of entries in the list.
	Count                float32       `json:"count"`
	Items                []GetGroupDto `json:"items"`
	AdditionalProperties map[string]interface{}
}

type _GroupListDto GroupListDto

// NewGroupListDto instantiates a new GroupListDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGroupListDto(count float32, items []GetGroupDto) *GroupListDto {
	this := GroupListDto{}
	this.Count = count
	this.Items = items
	return &this
}

// NewGroupListDtoWithDefaults instantiates a new GroupListDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGroupListDtoWithDefaults() *GroupListDto {
	this := GroupListDto{}
	return &this
}

// GetCount returns the Count field value
func (o *GroupListDto) GetCount() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.Count
}

// GetCountOk returns a tuple with the Count field value
// and a boolean to check if the value has been set.
func (o *GroupListDto) GetCountOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Count, true
}

// SetCount sets field value
func (o *GroupListDto) SetCount(v float32) {
	o.Count = v
}

// GetItems returns the Items field value
func (o *GroupListDto) GetItems() []GetGroupDto {
	if o == nil {
		var ret []GetGroupDto
		return ret
	}

	return o.Items
}

// GetItemsOk returns a tuple with the Items field value
// and a boolean to check if the value has been set.
func (o *GroupListDto) GetItemsOk() ([]GetGroupDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Items, true
}

// SetItems sets field value
func (o *GroupListDto) SetItems(v []GetGroupDto) {
	o.Items = v
}

func (o GroupListDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GroupListDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["count"] = o.Count
	toSerialize["items"] = o.Items

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *GroupListDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"count",
		"items",
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

	varGroupListDto := _GroupListDto{}

	err = json.Unmarshal(data, &varGroupListDto)

	if err != nil {
		return err
	}

	*o = GroupListDto(varGroupListDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "count")
		delete(additionalProperties, "items")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableGroupListDto struct {
	value *GroupListDto
	isSet bool
}

func (v NullableGroupListDto) Get() *GroupListDto {
	return v.value
}

func (v *NullableGroupListDto) Set(val *GroupListDto) {
	v.value = val
	v.isSet = true
}

func (v NullableGroupListDto) IsSet() bool {
	return v.isSet
}

func (v *NullableGroupListDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGroupListDto(val *GroupListDto) *NullableGroupListDto {
	return &NullableGroupListDto{value: val, isSet: true}
}

func (v NullableGroupListDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGroupListDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
