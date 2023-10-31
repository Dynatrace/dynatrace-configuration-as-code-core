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

// checks if the PolicyUuidsDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PolicyUuidsDto{}

// PolicyUuidsDto struct for PolicyUuidsDto
type PolicyUuidsDto struct {
	// A list of policies bound to the user group.
	PolicyUuids []string `json:"policyUuids"`
}

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
	return toSerialize, nil
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
