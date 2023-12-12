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

// RulePreviewRequest - struct for RulePreviewRequest
type RulePreviewRequest struct {
	FixedOffsetRulePreviewRequest    *FixedOffsetRulePreviewRequest
	GroupingRulePreviewRequest       *GroupingRulePreviewRequest
	RRulePreviewRequest              *RRulePreviewRequest
	RelativeOffsetRulePreviewRequest *RelativeOffsetRulePreviewRequest
}

// FixedOffsetRulePreviewRequestAsRulePreviewRequest is a convenience function that returns FixedOffsetRulePreviewRequest wrapped in RulePreviewRequest
func FixedOffsetRulePreviewRequestAsRulePreviewRequest(v *FixedOffsetRulePreviewRequest) RulePreviewRequest {
	return RulePreviewRequest{
		FixedOffsetRulePreviewRequest: v,
	}
}

// GroupingRulePreviewRequestAsRulePreviewRequest is a convenience function that returns GroupingRulePreviewRequest wrapped in RulePreviewRequest
func GroupingRulePreviewRequestAsRulePreviewRequest(v *GroupingRulePreviewRequest) RulePreviewRequest {
	return RulePreviewRequest{
		GroupingRulePreviewRequest: v,
	}
}

// RRulePreviewRequestAsRulePreviewRequest is a convenience function that returns RRulePreviewRequest wrapped in RulePreviewRequest
func RRulePreviewRequestAsRulePreviewRequest(v *RRulePreviewRequest) RulePreviewRequest {
	return RulePreviewRequest{
		RRulePreviewRequest: v,
	}
}

// RelativeOffsetRulePreviewRequestAsRulePreviewRequest is a convenience function that returns RelativeOffsetRulePreviewRequest wrapped in RulePreviewRequest
func RelativeOffsetRulePreviewRequestAsRulePreviewRequest(v *RelativeOffsetRulePreviewRequest) RulePreviewRequest {
	return RulePreviewRequest{
		RelativeOffsetRulePreviewRequest: v,
	}
}

// Unmarshal JSON data into one of the pointers in the struct
func (dst *RulePreviewRequest) UnmarshalJSON(data []byte) error {
	var err error
	match := 0
	// try to unmarshal data into FixedOffsetRulePreviewRequest
	err = newStrictDecoder(data).Decode(&dst.FixedOffsetRulePreviewRequest)
	if err == nil {
		jsonFixedOffsetRulePreviewRequest, _ := json.Marshal(dst.FixedOffsetRulePreviewRequest)
		if string(jsonFixedOffsetRulePreviewRequest) == "{}" { // empty struct
			dst.FixedOffsetRulePreviewRequest = nil
		} else {
			match++
		}
	} else {
		dst.FixedOffsetRulePreviewRequest = nil
	}

	// try to unmarshal data into GroupingRulePreviewRequest
	err = newStrictDecoder(data).Decode(&dst.GroupingRulePreviewRequest)
	if err == nil {
		jsonGroupingRulePreviewRequest, _ := json.Marshal(dst.GroupingRulePreviewRequest)
		if string(jsonGroupingRulePreviewRequest) == "{}" { // empty struct
			dst.GroupingRulePreviewRequest = nil
		} else {
			match++
		}
	} else {
		dst.GroupingRulePreviewRequest = nil
	}

	// try to unmarshal data into RRulePreviewRequest
	err = newStrictDecoder(data).Decode(&dst.RRulePreviewRequest)
	if err == nil {
		jsonRRulePreviewRequest, _ := json.Marshal(dst.RRulePreviewRequest)
		if string(jsonRRulePreviewRequest) == "{}" { // empty struct
			dst.RRulePreviewRequest = nil
		} else {
			match++
		}
	} else {
		dst.RRulePreviewRequest = nil
	}

	// try to unmarshal data into RelativeOffsetRulePreviewRequest
	err = newStrictDecoder(data).Decode(&dst.RelativeOffsetRulePreviewRequest)
	if err == nil {
		jsonRelativeOffsetRulePreviewRequest, _ := json.Marshal(dst.RelativeOffsetRulePreviewRequest)
		if string(jsonRelativeOffsetRulePreviewRequest) == "{}" { // empty struct
			dst.RelativeOffsetRulePreviewRequest = nil
		} else {
			match++
		}
	} else {
		dst.RelativeOffsetRulePreviewRequest = nil
	}

	if match > 1 { // more than 1 match
		// reset to nil
		dst.FixedOffsetRulePreviewRequest = nil
		dst.GroupingRulePreviewRequest = nil
		dst.RRulePreviewRequest = nil
		dst.RelativeOffsetRulePreviewRequest = nil

		return fmt.Errorf("data matches more than one schema in oneOf(RulePreviewRequest)")
	} else if match == 1 {
		return nil // exactly one match
	} else { // no match
		return fmt.Errorf("data failed to match schemas in oneOf(RulePreviewRequest)")
	}
}

// Marshal data from the first non-nil pointers in the struct to JSON
func (src RulePreviewRequest) MarshalJSON() ([]byte, error) {
	if src.FixedOffsetRulePreviewRequest != nil {
		return json.Marshal(&src.FixedOffsetRulePreviewRequest)
	}

	if src.GroupingRulePreviewRequest != nil {
		return json.Marshal(&src.GroupingRulePreviewRequest)
	}

	if src.RRulePreviewRequest != nil {
		return json.Marshal(&src.RRulePreviewRequest)
	}

	if src.RelativeOffsetRulePreviewRequest != nil {
		return json.Marshal(&src.RelativeOffsetRulePreviewRequest)
	}

	return nil, nil // no data in oneOf schemas
}

// Get the actual instance
func (obj *RulePreviewRequest) GetActualInstance() interface{} {
	if obj == nil {
		return nil
	}
	if obj.FixedOffsetRulePreviewRequest != nil {
		return obj.FixedOffsetRulePreviewRequest
	}

	if obj.GroupingRulePreviewRequest != nil {
		return obj.GroupingRulePreviewRequest
	}

	if obj.RRulePreviewRequest != nil {
		return obj.RRulePreviewRequest
	}

	if obj.RelativeOffsetRulePreviewRequest != nil {
		return obj.RelativeOffsetRulePreviewRequest
	}

	// all schemas are nil
	return nil
}

type NullableRulePreviewRequest struct {
	value *RulePreviewRequest
	isSet bool
}

func (v NullableRulePreviewRequest) Get() *RulePreviewRequest {
	return v.value
}

func (v *NullableRulePreviewRequest) Set(val *RulePreviewRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableRulePreviewRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableRulePreviewRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableRulePreviewRequest(val *RulePreviewRequest) *NullableRulePreviewRequest {
	return &NullableRulePreviewRequest{value: val, isSet: true}
}

func (v NullableRulePreviewRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableRulePreviewRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
