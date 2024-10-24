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

// checks if the PolicyUuidsDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PolicyUuidsDto{}

// PolicyUuidsDto struct for PolicyUuidsDto
type PolicyUuidsDto struct {
	// A list of policies bound to the user group.
	PolicyUuids          []string `json:"policyUuids"`
	AdditionalProperties map[string]interface{}
}

type _PolicyUuidsDto PolicyUuidsDto

// NewPolicyUuidsDto instantiates a new PolicyUuidsDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPolicyUuidsDto(policyUuids []string) *PolicyUuidsDto {
	this := PolicyUuidsDto{}
	this.PolicyUuids = policyUuids
	return &this
}

// NewPolicyUuidsDtoWithDefaults instantiates a new PolicyUuidsDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPolicyUuidsDtoWithDefaults() *PolicyUuidsDto {
	this := PolicyUuidsDto{}
	return &this
}

// GetPolicyUuids returns the PolicyUuids field value
func (o *PolicyUuidsDto) GetPolicyUuids() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.PolicyUuids
}

// GetPolicyUuidsOk returns a tuple with the PolicyUuids field value
// and a boolean to check if the value has been set.
func (o *PolicyUuidsDto) GetPolicyUuidsOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.PolicyUuids, true
}

// SetPolicyUuids sets field value
func (o *PolicyUuidsDto) SetPolicyUuids(v []string) {
	o.PolicyUuids = v
}

func (o PolicyUuidsDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PolicyUuidsDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["policyUuids"] = o.PolicyUuids

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *PolicyUuidsDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"policyUuids",
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

	varPolicyUuidsDto := _PolicyUuidsDto{}

	err = json.Unmarshal(data, &varPolicyUuidsDto)

	if err != nil {
		return err
	}

	*o = PolicyUuidsDto(varPolicyUuidsDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "policyUuids")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullablePolicyUuidsDto struct {
	value *PolicyUuidsDto
	isSet bool
}

func (v NullablePolicyUuidsDto) Get() *PolicyUuidsDto {
	return v.value
}

func (v *NullablePolicyUuidsDto) Set(val *PolicyUuidsDto) {
	v.value = val
	v.isSet = true
}

func (v NullablePolicyUuidsDto) IsSet() bool {
	return v.isSet
}

func (v *NullablePolicyUuidsDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePolicyUuidsDto(val *PolicyUuidsDto) *NullablePolicyUuidsDto {
	return &NullablePolicyUuidsDto{value: val, isSet: true}
}

func (v NullablePolicyUuidsDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePolicyUuidsDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
