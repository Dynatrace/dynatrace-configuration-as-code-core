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

// checks if the LevelPolicyBindingDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &LevelPolicyBindingDto{}

// LevelPolicyBindingDto struct for LevelPolicyBindingDto
type LevelPolicyBindingDto struct {
	// The type of the policy level.
	LevelType string `json:"levelType"`
	// The ID of the policy level.
	LevelId              string    `json:"levelId"`
	PolicyBindings       []Binding `json:"policyBindings"`
	AdditionalProperties map[string]interface{}
}

type _LevelPolicyBindingDto LevelPolicyBindingDto

// NewLevelPolicyBindingDto instantiates a new LevelPolicyBindingDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLevelPolicyBindingDto(levelType string, levelId string, policyBindings []Binding) *LevelPolicyBindingDto {
	this := LevelPolicyBindingDto{}
	this.LevelType = levelType
	this.LevelId = levelId
	this.PolicyBindings = policyBindings
	return &this
}

// NewLevelPolicyBindingDtoWithDefaults instantiates a new LevelPolicyBindingDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLevelPolicyBindingDtoWithDefaults() *LevelPolicyBindingDto {
	this := LevelPolicyBindingDto{}
	return &this
}

// GetLevelType returns the LevelType field value
func (o *LevelPolicyBindingDto) GetLevelType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.LevelType
}

// GetLevelTypeOk returns a tuple with the LevelType field value
// and a boolean to check if the value has been set.
func (o *LevelPolicyBindingDto) GetLevelTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LevelType, true
}

// SetLevelType sets field value
func (o *LevelPolicyBindingDto) SetLevelType(v string) {
	o.LevelType = v
}

// GetLevelId returns the LevelId field value
func (o *LevelPolicyBindingDto) GetLevelId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.LevelId
}

// GetLevelIdOk returns a tuple with the LevelId field value
// and a boolean to check if the value has been set.
func (o *LevelPolicyBindingDto) GetLevelIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LevelId, true
}

// SetLevelId sets field value
func (o *LevelPolicyBindingDto) SetLevelId(v string) {
	o.LevelId = v
}

// GetPolicyBindings returns the PolicyBindings field value
func (o *LevelPolicyBindingDto) GetPolicyBindings() []Binding {
	if o == nil {
		var ret []Binding
		return ret
	}

	return o.PolicyBindings
}

// GetPolicyBindingsOk returns a tuple with the PolicyBindings field value
// and a boolean to check if the value has been set.
func (o *LevelPolicyBindingDto) GetPolicyBindingsOk() ([]Binding, bool) {
	if o == nil {
		return nil, false
	}
	return o.PolicyBindings, true
}

// SetPolicyBindings sets field value
func (o *LevelPolicyBindingDto) SetPolicyBindings(v []Binding) {
	o.PolicyBindings = v
}

func (o LevelPolicyBindingDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o LevelPolicyBindingDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["levelType"] = o.LevelType
	toSerialize["levelId"] = o.LevelId
	toSerialize["policyBindings"] = o.PolicyBindings

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *LevelPolicyBindingDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"levelType",
		"levelId",
		"policyBindings",
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

	varLevelPolicyBindingDto := _LevelPolicyBindingDto{}

	err = json.Unmarshal(data, &varLevelPolicyBindingDto)

	if err != nil {
		return err
	}

	*o = LevelPolicyBindingDto(varLevelPolicyBindingDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "levelType")
		delete(additionalProperties, "levelId")
		delete(additionalProperties, "policyBindings")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableLevelPolicyBindingDto struct {
	value *LevelPolicyBindingDto
	isSet bool
}

func (v NullableLevelPolicyBindingDto) Get() *LevelPolicyBindingDto {
	return v.value
}

func (v *NullableLevelPolicyBindingDto) Set(val *LevelPolicyBindingDto) {
	v.value = val
	v.isSet = true
}

func (v NullableLevelPolicyBindingDto) IsSet() bool {
	return v.isSet
}

func (v *NullableLevelPolicyBindingDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLevelPolicyBindingDto(val *LevelPolicyBindingDto) *NullableLevelPolicyBindingDto {
	return &NullableLevelPolicyBindingDto{value: val, isSet: true}
}

func (v NullableLevelPolicyBindingDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLevelPolicyBindingDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
