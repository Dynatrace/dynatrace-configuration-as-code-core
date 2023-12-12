/*
Automation

Automation API allows working with workflows and various trigger options.

API version: 1.464.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package automation

import (
	"encoding/json"
)

// checks if the FixedOffsetRule type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &FixedOffsetRule{}

// FixedOffsetRule struct for FixedOffsetRule
type FixedOffsetRule struct {
	Rule string `json:"rule"`
	// Offset days
	Offset float32 `json:"offset"`
}

// NewFixedOffsetRule instantiates a new FixedOffsetRule object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewFixedOffsetRule(rule string, offset float32) *FixedOffsetRule {
	this := FixedOffsetRule{}
	this.Rule = rule
	this.Offset = offset
	return &this
}

// NewFixedOffsetRuleWithDefaults instantiates a new FixedOffsetRule object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewFixedOffsetRuleWithDefaults() *FixedOffsetRule {
	this := FixedOffsetRule{}
	return &this
}

// GetRule returns the Rule field value
func (o *FixedOffsetRule) GetRule() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Rule
}

// GetRuleOk returns a tuple with the Rule field value
// and a boolean to check if the value has been set.
func (o *FixedOffsetRule) GetRuleOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Rule, true
}

// SetRule sets field value
func (o *FixedOffsetRule) SetRule(v string) {
	o.Rule = v
}

// GetOffset returns the Offset field value
func (o *FixedOffsetRule) GetOffset() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.Offset
}

// GetOffsetOk returns a tuple with the Offset field value
// and a boolean to check if the value has been set.
func (o *FixedOffsetRule) GetOffsetOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Offset, true
}

// SetOffset sets field value
func (o *FixedOffsetRule) SetOffset(v float32) {
	o.Offset = v
}

func (o FixedOffsetRule) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o FixedOffsetRule) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["rule"] = o.Rule
	toSerialize["offset"] = o.Offset
	return toSerialize, nil
}

type NullableFixedOffsetRule struct {
	value *FixedOffsetRule
	isSet bool
}

func (v NullableFixedOffsetRule) Get() *FixedOffsetRule {
	return v.value
}

func (v *NullableFixedOffsetRule) Set(val *FixedOffsetRule) {
	v.value = val
	v.isSet = true
}

func (v NullableFixedOffsetRule) IsSet() bool {
	return v.isSet
}

func (v *NullableFixedOffsetRule) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableFixedOffsetRule(val *FixedOffsetRule) *NullableFixedOffsetRule {
	return &NullableFixedOffsetRule{value: val, isSet: true}
}

func (v NullableFixedOffsetRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableFixedOffsetRule) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
