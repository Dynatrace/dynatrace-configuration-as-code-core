/*
Automation

Automation API allows working with workflows and various trigger options.

API version: 1.464.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package automation

import (
	"encoding/json"
	"fmt"
)

// TaskConcurrency struct for TaskConcurrency
type TaskConcurrency struct {
	int32  *int32
	string *string
}

// Unmarshal JSON data into any of the pointers in the struct
func (dst *TaskConcurrency) UnmarshalJSON(data []byte) error {
	var err error
	// try to unmarshal JSON data into int32
	err = json.Unmarshal(data, &dst.int32)
	if err == nil {
		jsonint32, _ := json.Marshal(dst.int32)
		if string(jsonint32) == "{}" { // empty struct
			dst.int32 = nil
		} else {
			return nil // data stored in dst.int32, return on the first match
		}
	} else {
		dst.int32 = nil
	}

	// try to unmarshal JSON data into string
	err = json.Unmarshal(data, &dst.string)
	if err == nil {
		jsonstring, _ := json.Marshal(dst.string)
		if string(jsonstring) == "{}" { // empty struct
			dst.string = nil
		} else {
			return nil // data stored in dst.string, return on the first match
		}
	} else {
		dst.string = nil
	}

	return fmt.Errorf("data failed to match schemas in anyOf(TaskConcurrency)")
}

// Marshal data from the first non-nil pointers in the struct to JSON
func (src *TaskConcurrency) MarshalJSON() ([]byte, error) {
	if src.int32 != nil {
		return json.Marshal(&src.int32)
	}

	if src.string != nil {
		return json.Marshal(&src.string)
	}

	return nil, nil // no data in anyOf schemas
}

type NullableTaskConcurrency struct {
	value *TaskConcurrency
	isSet bool
}

func (v NullableTaskConcurrency) Get() *TaskConcurrency {
	return v.value
}

func (v *NullableTaskConcurrency) Set(val *TaskConcurrency) {
	v.value = val
	v.isSet = true
}

func (v NullableTaskConcurrency) IsSet() bool {
	return v.isSet
}

func (v *NullableTaskConcurrency) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableTaskConcurrency(val *TaskConcurrency) *NullableTaskConcurrency {
	return &NullableTaskConcurrency{value: val, isSet: true}
}

func (v NullableTaskConcurrency) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableTaskConcurrency) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
