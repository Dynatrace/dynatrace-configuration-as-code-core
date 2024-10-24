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

// checks if the EffectivePermission type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &EffectivePermission{}

// EffectivePermission struct for EffectivePermission
type EffectivePermission struct {
	// One of a effective permissions
	Permission string `json:"permission"`
	// A list of policies.
	Effects              []EffectivePermissionEffects `json:"effects"`
	AdditionalProperties map[string]interface{}
}

type _EffectivePermission EffectivePermission

// NewEffectivePermission instantiates a new EffectivePermission object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewEffectivePermission(permission string, effects []EffectivePermissionEffects) *EffectivePermission {
	this := EffectivePermission{}
	this.Permission = permission
	this.Effects = effects
	return &this
}

// NewEffectivePermissionWithDefaults instantiates a new EffectivePermission object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewEffectivePermissionWithDefaults() *EffectivePermission {
	this := EffectivePermission{}
	return &this
}

// GetPermission returns the Permission field value
func (o *EffectivePermission) GetPermission() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Permission
}

// GetPermissionOk returns a tuple with the Permission field value
// and a boolean to check if the value has been set.
func (o *EffectivePermission) GetPermissionOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Permission, true
}

// SetPermission sets field value
func (o *EffectivePermission) SetPermission(v string) {
	o.Permission = v
}

// GetEffects returns the Effects field value
func (o *EffectivePermission) GetEffects() []EffectivePermissionEffects {
	if o == nil {
		var ret []EffectivePermissionEffects
		return ret
	}

	return o.Effects
}

// GetEffectsOk returns a tuple with the Effects field value
// and a boolean to check if the value has been set.
func (o *EffectivePermission) GetEffectsOk() ([]EffectivePermissionEffects, bool) {
	if o == nil {
		return nil, false
	}
	return o.Effects, true
}

// SetEffects sets field value
func (o *EffectivePermission) SetEffects(v []EffectivePermissionEffects) {
	o.Effects = v
}

func (o EffectivePermission) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o EffectivePermission) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["permission"] = o.Permission
	toSerialize["effects"] = o.Effects

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *EffectivePermission) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"permission",
		"effects",
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

	varEffectivePermission := _EffectivePermission{}

	err = json.Unmarshal(data, &varEffectivePermission)

	if err != nil {
		return err
	}

	*o = EffectivePermission(varEffectivePermission)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "permission")
		delete(additionalProperties, "effects")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableEffectivePermission struct {
	value *EffectivePermission
	isSet bool
}

func (v NullableEffectivePermission) Get() *EffectivePermission {
	return v.value
}

func (v *NullableEffectivePermission) Set(val *EffectivePermission) {
	v.value = val
	v.isSet = true
}

func (v NullableEffectivePermission) IsSet() bool {
	return v.isSet
}

func (v *NullableEffectivePermission) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEffectivePermission(val *EffectivePermission) *NullableEffectivePermission {
	return &NullableEffectivePermission{value: val, isSet: true}
}

func (v NullableEffectivePermission) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEffectivePermission) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
