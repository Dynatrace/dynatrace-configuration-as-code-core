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

// checks if the ServiceUserNameDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ServiceUserNameDto{}

// ServiceUserNameDto struct for ServiceUserNameDto
type ServiceUserNameDto struct {
	// The name of the new service user
	Name string `json:"name"`
}

// NewServiceUserNameDto instantiates a new ServiceUserNameDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewServiceUserNameDto(name string) *ServiceUserNameDto {
	this := ServiceUserNameDto{}
	this.Name = name
	return &this
}

// NewServiceUserNameDtoWithDefaults instantiates a new ServiceUserNameDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewServiceUserNameDtoWithDefaults() *ServiceUserNameDto {
	this := ServiceUserNameDto{}
	return &this
}

// GetName returns the Name field value
func (o *ServiceUserNameDto) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *ServiceUserNameDto) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *ServiceUserNameDto) SetName(v string) {
	o.Name = v
}

func (o ServiceUserNameDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ServiceUserNameDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["name"] = o.Name
	return toSerialize, nil
}

type NullableServiceUserNameDto struct {
	value *ServiceUserNameDto
	isSet bool
}

func (v NullableServiceUserNameDto) Get() *ServiceUserNameDto {
	return v.value
}

func (v *NullableServiceUserNameDto) Set(val *ServiceUserNameDto) {
	v.value = val
	v.isSet = true
}

func (v NullableServiceUserNameDto) IsSet() bool {
	return v.isSet
}

func (v *NullableServiceUserNameDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableServiceUserNameDto(val *ServiceUserNameDto) *NullableServiceUserNameDto {
	return &NullableServiceUserNameDto{value: val, isSet: true}
}

func (v NullableServiceUserNameDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableServiceUserNameDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
