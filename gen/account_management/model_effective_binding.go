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

// checks if the EffectiveBinding type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &EffectiveBinding{}

// EffectiveBinding struct for EffectiveBinding
type EffectiveBinding struct {
	// The UUID of group
	GroupUuid string `json:"groupUuid"`
	// The type of the level to which the binding applies.
	LevelType string `json:"levelType"`
	// The ID of the level to which the binding applies.
	LevelId              string `json:"levelId"`
	AdditionalProperties map[string]interface{}
}

type _EffectiveBinding EffectiveBinding

// NewEffectiveBinding instantiates a new EffectiveBinding object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewEffectiveBinding(groupUuid string, levelType string, levelId string) *EffectiveBinding {
	this := EffectiveBinding{}
	this.GroupUuid = groupUuid
	this.LevelType = levelType
	this.LevelId = levelId
	return &this
}

// NewEffectiveBindingWithDefaults instantiates a new EffectiveBinding object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewEffectiveBindingWithDefaults() *EffectiveBinding {
	this := EffectiveBinding{}
	return &this
}

// GetGroupUuid returns the GroupUuid field value
func (o *EffectiveBinding) GetGroupUuid() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.GroupUuid
}

// GetGroupUuidOk returns a tuple with the GroupUuid field value
// and a boolean to check if the value has been set.
func (o *EffectiveBinding) GetGroupUuidOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.GroupUuid, true
}

// SetGroupUuid sets field value
func (o *EffectiveBinding) SetGroupUuid(v string) {
	o.GroupUuid = v
}

// GetLevelType returns the LevelType field value
func (o *EffectiveBinding) GetLevelType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.LevelType
}

// GetLevelTypeOk returns a tuple with the LevelType field value
// and a boolean to check if the value has been set.
func (o *EffectiveBinding) GetLevelTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LevelType, true
}

// SetLevelType sets field value
func (o *EffectiveBinding) SetLevelType(v string) {
	o.LevelType = v
}

// GetLevelId returns the LevelId field value
func (o *EffectiveBinding) GetLevelId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.LevelId
}

// GetLevelIdOk returns a tuple with the LevelId field value
// and a boolean to check if the value has been set.
func (o *EffectiveBinding) GetLevelIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LevelId, true
}

// SetLevelId sets field value
func (o *EffectiveBinding) SetLevelId(v string) {
	o.LevelId = v
}

func (o EffectiveBinding) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o EffectiveBinding) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["groupUuid"] = o.GroupUuid
	toSerialize["levelType"] = o.LevelType
	toSerialize["levelId"] = o.LevelId

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *EffectiveBinding) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"groupUuid",
		"levelType",
		"levelId",
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

	varEffectiveBinding := _EffectiveBinding{}

	err = json.Unmarshal(data, &varEffectiveBinding)

	if err != nil {
		return err
	}

	*o = EffectiveBinding(varEffectiveBinding)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "groupUuid")
		delete(additionalProperties, "levelType")
		delete(additionalProperties, "levelId")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableEffectiveBinding struct {
	value *EffectiveBinding
	isSet bool
}

func (v NullableEffectiveBinding) Get() *EffectiveBinding {
	return v.value
}

func (v *NullableEffectiveBinding) Set(val *EffectiveBinding) {
	v.value = val
	v.isSet = true
}

func (v NullableEffectiveBinding) IsSet() bool {
	return v.isSet
}

func (v *NullableEffectiveBinding) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEffectiveBinding(val *EffectiveBinding) *NullableEffectiveBinding {
	return &NullableEffectiveBinding{value: val, isSet: true}
}

func (v NullableEffectiveBinding) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEffectiveBinding) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
