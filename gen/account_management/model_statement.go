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

// checks if the Statement type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Statement{}

// Statement struct for Statement
type Statement struct {
	// The effect of the policy (for example, allow something).
	Effect string `json:"effect"`
	// A list of granted permissions.
	Permissions []string `json:"permissions"`
	// A list of conditions limiting the granted permissions.
	Conditions           []Condition `json:"conditions"`
	AdditionalProperties map[string]interface{}
}

type _Statement Statement

// NewStatement instantiates a new Statement object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewStatement(effect string, permissions []string, conditions []Condition) *Statement {
	this := Statement{}
	this.Effect = effect
	this.Permissions = permissions
	this.Conditions = conditions
	return &this
}

// NewStatementWithDefaults instantiates a new Statement object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewStatementWithDefaults() *Statement {
	this := Statement{}
	return &this
}

// GetEffect returns the Effect field value
func (o *Statement) GetEffect() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Effect
}

// GetEffectOk returns a tuple with the Effect field value
// and a boolean to check if the value has been set.
func (o *Statement) GetEffectOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Effect, true
}

// SetEffect sets field value
func (o *Statement) SetEffect(v string) {
	o.Effect = v
}

// GetPermissions returns the Permissions field value
func (o *Statement) GetPermissions() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.Permissions
}

// GetPermissionsOk returns a tuple with the Permissions field value
// and a boolean to check if the value has been set.
func (o *Statement) GetPermissionsOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.Permissions, true
}

// SetPermissions sets field value
func (o *Statement) SetPermissions(v []string) {
	o.Permissions = v
}

// GetConditions returns the Conditions field value
func (o *Statement) GetConditions() []Condition {
	if o == nil {
		var ret []Condition
		return ret
	}

	return o.Conditions
}

// GetConditionsOk returns a tuple with the Conditions field value
// and a boolean to check if the value has been set.
func (o *Statement) GetConditionsOk() ([]Condition, bool) {
	if o == nil {
		return nil, false
	}
	return o.Conditions, true
}

// SetConditions sets field value
func (o *Statement) SetConditions(v []Condition) {
	o.Conditions = v
}

func (o Statement) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Statement) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["effect"] = o.Effect
	toSerialize["permissions"] = o.Permissions
	toSerialize["conditions"] = o.Conditions

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *Statement) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"effect",
		"permissions",
		"conditions",
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

	varStatement := _Statement{}

	err = json.Unmarshal(data, &varStatement)

	if err != nil {
		return err
	}

	*o = Statement(varStatement)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "effect")
		delete(additionalProperties, "permissions")
		delete(additionalProperties, "conditions")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableStatement struct {
	value *Statement
	isSet bool
}

func (v NullableStatement) Get() *Statement {
	return v.value
}

func (v *NullableStatement) Set(val *Statement) {
	v.value = val
	v.isSet = true
}

func (v NullableStatement) IsSet() bool {
	return v.isSet
}

func (v *NullableStatement) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableStatement(val *Statement) *NullableStatement {
	return &NullableStatement{value: val, isSet: true}
}

func (v NullableStatement) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableStatement) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
