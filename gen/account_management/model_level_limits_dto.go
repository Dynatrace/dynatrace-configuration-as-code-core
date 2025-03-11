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

// checks if the LevelLimitsDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &LevelLimitsDto{}

// LevelLimitsDto struct for LevelLimitsDto
type LevelLimitsDto struct {
	// Information about policies limit set for a level.
	Policies LimitEntry `json:"policies"`
	// Information about policy bindings limit set for a level.
	Bindings LimitEntry `json:"bindings"`
	// Information about policy boundaries limit set for a level.
	Boundaries           LimitEntry `json:"boundaries"`
	AdditionalProperties map[string]interface{}
}

type _LevelLimitsDto LevelLimitsDto

// NewLevelLimitsDto instantiates a new LevelLimitsDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewLevelLimitsDto(policies LimitEntry, bindings LimitEntry, boundaries LimitEntry) *LevelLimitsDto {
	this := LevelLimitsDto{}
	this.Policies = policies
	this.Bindings = bindings
	this.Boundaries = boundaries
	return &this
}

// NewLevelLimitsDtoWithDefaults instantiates a new LevelLimitsDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewLevelLimitsDtoWithDefaults() *LevelLimitsDto {
	this := LevelLimitsDto{}
	return &this
}

// GetPolicies returns the Policies field value
func (o *LevelLimitsDto) GetPolicies() LimitEntry {
	if o == nil {
		var ret LimitEntry
		return ret
	}

	return o.Policies
}

// GetPoliciesOk returns a tuple with the Policies field value
// and a boolean to check if the value has been set.
func (o *LevelLimitsDto) GetPoliciesOk() (*LimitEntry, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Policies, true
}

// SetPolicies sets field value
func (o *LevelLimitsDto) SetPolicies(v LimitEntry) {
	o.Policies = v
}

// GetBindings returns the Bindings field value
func (o *LevelLimitsDto) GetBindings() LimitEntry {
	if o == nil {
		var ret LimitEntry
		return ret
	}

	return o.Bindings
}

// GetBindingsOk returns a tuple with the Bindings field value
// and a boolean to check if the value has been set.
func (o *LevelLimitsDto) GetBindingsOk() (*LimitEntry, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Bindings, true
}

// SetBindings sets field value
func (o *LevelLimitsDto) SetBindings(v LimitEntry) {
	o.Bindings = v
}

// GetBoundaries returns the Boundaries field value
func (o *LevelLimitsDto) GetBoundaries() LimitEntry {
	if o == nil {
		var ret LimitEntry
		return ret
	}

	return o.Boundaries
}

// GetBoundariesOk returns a tuple with the Boundaries field value
// and a boolean to check if the value has been set.
func (o *LevelLimitsDto) GetBoundariesOk() (*LimitEntry, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Boundaries, true
}

// SetBoundaries sets field value
func (o *LevelLimitsDto) SetBoundaries(v LimitEntry) {
	o.Boundaries = v
}

func (o LevelLimitsDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o LevelLimitsDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["policies"] = o.Policies
	toSerialize["bindings"] = o.Bindings
	toSerialize["boundaries"] = o.Boundaries

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *LevelLimitsDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"policies",
		"bindings",
		"boundaries",
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

	varLevelLimitsDto := _LevelLimitsDto{}

	err = json.Unmarshal(data, &varLevelLimitsDto)

	if err != nil {
		return err
	}

	*o = LevelLimitsDto(varLevelLimitsDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "policies")
		delete(additionalProperties, "bindings")
		delete(additionalProperties, "boundaries")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableLevelLimitsDto struct {
	value *LevelLimitsDto
	isSet bool
}

func (v NullableLevelLimitsDto) Get() *LevelLimitsDto {
	return v.value
}

func (v *NullableLevelLimitsDto) Set(val *LevelLimitsDto) {
	v.value = val
	v.isSet = true
}

func (v NullableLevelLimitsDto) IsSet() bool {
	return v.isSet
}

func (v *NullableLevelLimitsDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableLevelLimitsDto(val *LevelLimitsDto) *NullableLevelLimitsDto {
	return &NullableLevelLimitsDto{value: val, isSet: true}
}

func (v NullableLevelLimitsDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableLevelLimitsDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
