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

// checks if the GroupUserListDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GroupUserListDto{}

// GroupUserListDto struct for GroupUserListDto
type GroupUserListDto struct {
	// The number of entries in the list.
	Count float32   `json:"count"`
	Items []UserDto `json:"items"`
}

// NewGroupUserListDto instantiates a new GroupUserListDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGroupUserListDto(count float32, items []UserDto) *GroupUserListDto {
	this := GroupUserListDto{}
	this.Count = count
	this.Items = items
	return &this
}

// NewGroupUserListDtoWithDefaults instantiates a new GroupUserListDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGroupUserListDtoWithDefaults() *GroupUserListDto {
	this := GroupUserListDto{}
	return &this
}

// GetCount returns the Count field value
func (o *GroupUserListDto) GetCount() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.Count
}

// GetCountOk returns a tuple with the Count field value
// and a boolean to check if the value has been set.
func (o *GroupUserListDto) GetCountOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Count, true
}

// SetCount sets field value
func (o *GroupUserListDto) SetCount(v float32) {
	o.Count = v
}

// GetItems returns the Items field value
func (o *GroupUserListDto) GetItems() []UserDto {
	if o == nil {
		var ret []UserDto
		return ret
	}

	return o.Items
}

// GetItemsOk returns a tuple with the Items field value
// and a boolean to check if the value has been set.
func (o *GroupUserListDto) GetItemsOk() ([]UserDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Items, true
}

// SetItems sets field value
func (o *GroupUserListDto) SetItems(v []UserDto) {
	o.Items = v
}

func (o GroupUserListDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GroupUserListDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["count"] = o.Count
	toSerialize["items"] = o.Items
	return toSerialize, nil
}

type NullableGroupUserListDto struct {
	value *GroupUserListDto
	isSet bool
}

func (v NullableGroupUserListDto) Get() *GroupUserListDto {
	return v.value
}

func (v *NullableGroupUserListDto) Set(val *GroupUserListDto) {
	v.value = val
	v.isSet = true
}

func (v NullableGroupUserListDto) IsSet() bool {
	return v.isSet
}

func (v *NullableGroupUserListDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGroupUserListDto(val *GroupUserListDto) *NullableGroupUserListDto {
	return &NullableGroupUserListDto{value: val, isSet: true}
}

func (v NullableGroupUserListDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGroupUserListDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}