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

// checks if the TimeTrigger type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &TimeTrigger{}

// TimeTrigger struct for TimeTrigger
type TimeTrigger struct {
	Type string `json:"type"`
	Time string `json:"time"`
}

// NewTimeTrigger instantiates a new TimeTrigger object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTimeTrigger(type_ string, time string) *TimeTrigger {
	this := TimeTrigger{}
	this.Type = type_
	this.Time = time
	return &this
}

// NewTimeTriggerWithDefaults instantiates a new TimeTrigger object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTimeTriggerWithDefaults() *TimeTrigger {
	this := TimeTrigger{}
	return &this
}

// GetType returns the Type field value
func (o *TimeTrigger) GetType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Type
}

// GetTypeOk returns a tuple with the Type field value
// and a boolean to check if the value has been set.
func (o *TimeTrigger) GetTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Type, true
}

// SetType sets field value
func (o *TimeTrigger) SetType(v string) {
	o.Type = v
}

// GetTime returns the Time field value
func (o *TimeTrigger) GetTime() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Time
}

// GetTimeOk returns a tuple with the Time field value
// and a boolean to check if the value has been set.
func (o *TimeTrigger) GetTimeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Time, true
}

// SetTime sets field value
func (o *TimeTrigger) SetTime(v string) {
	o.Time = v
}

func (o TimeTrigger) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o TimeTrigger) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["type"] = o.Type
	toSerialize["time"] = o.Time
	return toSerialize, nil
}

type NullableTimeTrigger struct {
	value *TimeTrigger
	isSet bool
}

func (v NullableTimeTrigger) Get() *TimeTrigger {
	return v.value
}

func (v *NullableTimeTrigger) Set(val *TimeTrigger) {
	v.value = val
	v.isSet = true
}

func (v NullableTimeTrigger) IsSet() bool {
	return v.isSet
}

func (v *NullableTimeTrigger) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTimeTrigger(val *TimeTrigger) *NullableTimeTrigger {
	return &NullableTimeTrigger{value: val, isSet: true}
}

func (v NullableTimeTrigger) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTimeTrigger) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
