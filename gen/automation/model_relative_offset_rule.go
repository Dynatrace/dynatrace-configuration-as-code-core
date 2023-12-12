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

// checks if the RelativeOffsetRule type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &RelativeOffsetRule{}

// RelativeOffsetRule struct for RelativeOffsetRule
type RelativeOffsetRule struct {
	Direction  string `json:"direction"`
	SourceRule string `json:"sourceRule"`
	TargetRule string `json:"targetRule"`
}

// NewRelativeOffsetRule instantiates a new RelativeOffsetRule object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewRelativeOffsetRule(direction string, sourceRule string, targetRule string) *RelativeOffsetRule {
	this := RelativeOffsetRule{}
	this.Direction = direction
	this.SourceRule = sourceRule
	this.TargetRule = targetRule
	return &this
}

// NewRelativeOffsetRuleWithDefaults instantiates a new RelativeOffsetRule object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewRelativeOffsetRuleWithDefaults() *RelativeOffsetRule {
	this := RelativeOffsetRule{}
	return &this
}

// GetDirection returns the Direction field value
func (o *RelativeOffsetRule) GetDirection() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Direction
}

// GetDirectionOk returns a tuple with the Direction field value
// and a boolean to check if the value has been set.
func (o *RelativeOffsetRule) GetDirectionOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Direction, true
}

// SetDirection sets field value
func (o *RelativeOffsetRule) SetDirection(v string) {
	o.Direction = v
}

// GetSourceRule returns the SourceRule field value
func (o *RelativeOffsetRule) GetSourceRule() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.SourceRule
}

// GetSourceRuleOk returns a tuple with the SourceRule field value
// and a boolean to check if the value has been set.
func (o *RelativeOffsetRule) GetSourceRuleOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.SourceRule, true
}

// SetSourceRule sets field value
func (o *RelativeOffsetRule) SetSourceRule(v string) {
	o.SourceRule = v
}

// GetTargetRule returns the TargetRule field value
func (o *RelativeOffsetRule) GetTargetRule() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.TargetRule
}

// GetTargetRuleOk returns a tuple with the TargetRule field value
// and a boolean to check if the value has been set.
func (o *RelativeOffsetRule) GetTargetRuleOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TargetRule, true
}

// SetTargetRule sets field value
func (o *RelativeOffsetRule) SetTargetRule(v string) {
	o.TargetRule = v
}

func (o RelativeOffsetRule) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o RelativeOffsetRule) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["direction"] = o.Direction
	toSerialize["sourceRule"] = o.SourceRule
	toSerialize["targetRule"] = o.TargetRule
	return toSerialize, nil
}

type NullableRelativeOffsetRule struct {
	value *RelativeOffsetRule
	isSet bool
}

func (v NullableRelativeOffsetRule) Get() *RelativeOffsetRule {
	return v.value
}

func (v *NullableRelativeOffsetRule) Set(val *RelativeOffsetRule) {
	v.value = val
	v.isSet = true
}

func (v NullableRelativeOffsetRule) IsSet() bool {
	return v.isSet
}

func (v *NullableRelativeOffsetRule) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRelativeOffsetRule(val *RelativeOffsetRule) *NullableRelativeOffsetRule {
	return &NullableRelativeOffsetRule{value: val, isSet: true}
}

func (v NullableRelativeOffsetRule) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRelativeOffsetRule) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
