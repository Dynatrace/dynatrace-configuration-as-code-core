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

// checks if the TaskPosition type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &TaskPosition{}

// TaskPosition struct for TaskPosition
type TaskPosition struct {
	X int32 `json:"x"`
	Y int32 `json:"y"`
}

// NewTaskPosition instantiates a new TaskPosition object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewTaskPosition(x int32, y int32) *TaskPosition {
	this := TaskPosition{}
	this.X = x
	this.Y = y
	return &this
}

// NewTaskPositionWithDefaults instantiates a new TaskPosition object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewTaskPositionWithDefaults() *TaskPosition {
	this := TaskPosition{}
	return &this
}

// GetX returns the X field value
func (o *TaskPosition) GetX() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.X
}

// GetXOk returns a tuple with the X field value
// and a boolean to check if the value has been set.
func (o *TaskPosition) GetXOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.X, true
}

// SetX sets field value
func (o *TaskPosition) SetX(v int32) {
	o.X = v
}

// GetY returns the Y field value
func (o *TaskPosition) GetY() int32 {
	if o == nil {
		var ret int32
		return ret
	}

	return o.Y
}

// GetYOk returns a tuple with the Y field value
// and a boolean to check if the value has been set.
func (o *TaskPosition) GetYOk() (*int32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Y, true
}

// SetY sets field value
func (o *TaskPosition) SetY(v int32) {
	o.Y = v
}

func (o TaskPosition) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o TaskPosition) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["x"] = o.X
	toSerialize["y"] = o.Y
	return toSerialize, nil
}

type NullableTaskPosition struct {
	value *TaskPosition
	isSet bool
}

func (v NullableTaskPosition) Get() *TaskPosition {
	return v.value
}

func (v *NullableTaskPosition) Set(val *TaskPosition) {
	v.value = val
	v.isSet = true
}

func (v NullableTaskPosition) IsSet() bool {
	return v.isSet
}

func (v *NullableTaskPosition) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTaskPosition(val *TaskPosition) *NullableTaskPosition {
	return &NullableTaskPosition{value: val, isSet: true}
}

func (v NullableTaskPosition) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTaskPosition) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
