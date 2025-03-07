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

// checks if the PolicyDtoList type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PolicyDtoList{}

// PolicyDtoList struct for PolicyDtoList
type PolicyDtoList struct {
	// A list of policies.
	Policies             []PolicyDto `json:"policies"`
	AdditionalProperties map[string]interface{}
}

type _PolicyDtoList PolicyDtoList

// NewPolicyDtoList instantiates a new PolicyDtoList object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPolicyDtoList(policies []PolicyDto) *PolicyDtoList {
	this := PolicyDtoList{}
	this.Policies = policies
	return &this
}

// NewPolicyDtoListWithDefaults instantiates a new PolicyDtoList object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPolicyDtoListWithDefaults() *PolicyDtoList {
	this := PolicyDtoList{}
	return &this
}

// GetPolicies returns the Policies field value
func (o *PolicyDtoList) GetPolicies() []PolicyDto {
	if o == nil {
		var ret []PolicyDto
		return ret
	}

	return o.Policies
}

// GetPoliciesOk returns a tuple with the Policies field value
// and a boolean to check if the value has been set.
func (o *PolicyDtoList) GetPoliciesOk() ([]PolicyDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Policies, true
}

// SetPolicies sets field value
func (o *PolicyDtoList) SetPolicies(v []PolicyDto) {
	o.Policies = v
}

func (o PolicyDtoList) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PolicyDtoList) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["policies"] = o.Policies

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *PolicyDtoList) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"policies",
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

	varPolicyDtoList := _PolicyDtoList{}

	err = json.Unmarshal(data, &varPolicyDtoList)

	if err != nil {
		return err
	}

	*o = PolicyDtoList(varPolicyDtoList)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "policies")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullablePolicyDtoList struct {
	value *PolicyDtoList
	isSet bool
}

func (v NullablePolicyDtoList) Get() *PolicyDtoList {
	return v.value
}

func (v *NullablePolicyDtoList) Set(val *PolicyDtoList) {
	v.value = val
	v.isSet = true
}

func (v NullablePolicyDtoList) IsSet() bool {
	return v.isSet
}

func (v *NullablePolicyDtoList) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePolicyDtoList(val *PolicyDtoList) *NullablePolicyDtoList {
	return &NullablePolicyDtoList{value: val, isSet: true}
}

func (v NullablePolicyDtoList) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePolicyDtoList) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
