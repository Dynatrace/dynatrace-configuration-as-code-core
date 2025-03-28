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

// checks if the ClaEnvironmentDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ClaEnvironmentDto{}

// ClaEnvironmentDto struct for ClaEnvironmentDto
type ClaEnvironmentDto struct {
	// The ID of the environments.
	EnvironmentUuid      string `json:"environmentUuid"`
	AdditionalProperties map[string]interface{}
}

type _ClaEnvironmentDto ClaEnvironmentDto

// NewClaEnvironmentDto instantiates a new ClaEnvironmentDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewClaEnvironmentDto(environmentUuid string) *ClaEnvironmentDto {
	this := ClaEnvironmentDto{}
	this.EnvironmentUuid = environmentUuid
	return &this
}

// NewClaEnvironmentDtoWithDefaults instantiates a new ClaEnvironmentDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewClaEnvironmentDtoWithDefaults() *ClaEnvironmentDto {
	this := ClaEnvironmentDto{}
	return &this
}

// GetEnvironmentUuid returns the EnvironmentUuid field value
func (o *ClaEnvironmentDto) GetEnvironmentUuid() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.EnvironmentUuid
}

// GetEnvironmentUuidOk returns a tuple with the EnvironmentUuid field value
// and a boolean to check if the value has been set.
func (o *ClaEnvironmentDto) GetEnvironmentUuidOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EnvironmentUuid, true
}

// SetEnvironmentUuid sets field value
func (o *ClaEnvironmentDto) SetEnvironmentUuid(v string) {
	o.EnvironmentUuid = v
}

func (o ClaEnvironmentDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ClaEnvironmentDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["environmentUuid"] = o.EnvironmentUuid

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *ClaEnvironmentDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"environmentUuid",
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

	varClaEnvironmentDto := _ClaEnvironmentDto{}

	err = json.Unmarshal(data, &varClaEnvironmentDto)

	if err != nil {
		return err
	}

	*o = ClaEnvironmentDto(varClaEnvironmentDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "environmentUuid")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableClaEnvironmentDto struct {
	value *ClaEnvironmentDto
	isSet bool
}

func (v NullableClaEnvironmentDto) Get() *ClaEnvironmentDto {
	return v.value
}

func (v *NullableClaEnvironmentDto) Set(val *ClaEnvironmentDto) {
	v.value = val
	v.isSet = true
}

func (v NullableClaEnvironmentDto) IsSet() bool {
	return v.isSet
}

func (v *NullableClaEnvironmentDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableClaEnvironmentDto(val *ClaEnvironmentDto) *NullableClaEnvironmentDto {
	return &NullableClaEnvironmentDto{value: val, isSet: true}
}

func (v NullableClaEnvironmentDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableClaEnvironmentDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
