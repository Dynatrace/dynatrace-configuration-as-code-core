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

// checks if the ClaBudgetLimitRecordListDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ClaBudgetLimitRecordListDto{}

// ClaBudgetLimitRecordListDto struct for ClaBudgetLimitRecordListDto
type ClaBudgetLimitRecordListDto struct {
	// The number of entries in the list.
	TotalCount           float32                    `json:"totalCount"`
	Records              []ClaBudgetLimitRecordsDto `json:"records"`
	AdditionalProperties map[string]interface{}
}

type _ClaBudgetLimitRecordListDto ClaBudgetLimitRecordListDto

// NewClaBudgetLimitRecordListDto instantiates a new ClaBudgetLimitRecordListDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewClaBudgetLimitRecordListDto(totalCount float32, records []ClaBudgetLimitRecordsDto) *ClaBudgetLimitRecordListDto {
	this := ClaBudgetLimitRecordListDto{}
	this.TotalCount = totalCount
	this.Records = records
	return &this
}

// NewClaBudgetLimitRecordListDtoWithDefaults instantiates a new ClaBudgetLimitRecordListDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewClaBudgetLimitRecordListDtoWithDefaults() *ClaBudgetLimitRecordListDto {
	this := ClaBudgetLimitRecordListDto{}
	return &this
}

// GetTotalCount returns the TotalCount field value
func (o *ClaBudgetLimitRecordListDto) GetTotalCount() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.TotalCount
}

// GetTotalCountOk returns a tuple with the TotalCount field value
// and a boolean to check if the value has been set.
func (o *ClaBudgetLimitRecordListDto) GetTotalCountOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TotalCount, true
}

// SetTotalCount sets field value
func (o *ClaBudgetLimitRecordListDto) SetTotalCount(v float32) {
	o.TotalCount = v
}

// GetRecords returns the Records field value
func (o *ClaBudgetLimitRecordListDto) GetRecords() []ClaBudgetLimitRecordsDto {
	if o == nil {
		var ret []ClaBudgetLimitRecordsDto
		return ret
	}

	return o.Records
}

// GetRecordsOk returns a tuple with the Records field value
// and a boolean to check if the value has been set.
func (o *ClaBudgetLimitRecordListDto) GetRecordsOk() ([]ClaBudgetLimitRecordsDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Records, true
}

// SetRecords sets field value
func (o *ClaBudgetLimitRecordListDto) SetRecords(v []ClaBudgetLimitRecordsDto) {
	o.Records = v
}

func (o ClaBudgetLimitRecordListDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ClaBudgetLimitRecordListDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["totalCount"] = o.TotalCount
	toSerialize["records"] = o.Records

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *ClaBudgetLimitRecordListDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"totalCount",
		"records",
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

	varClaBudgetLimitRecordListDto := _ClaBudgetLimitRecordListDto{}

	err = json.Unmarshal(data, &varClaBudgetLimitRecordListDto)

	if err != nil {
		return err
	}

	*o = ClaBudgetLimitRecordListDto(varClaBudgetLimitRecordListDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "totalCount")
		delete(additionalProperties, "records")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableClaBudgetLimitRecordListDto struct {
	value *ClaBudgetLimitRecordListDto
	isSet bool
}

func (v NullableClaBudgetLimitRecordListDto) Get() *ClaBudgetLimitRecordListDto {
	return v.value
}

func (v *NullableClaBudgetLimitRecordListDto) Set(val *ClaBudgetLimitRecordListDto) {
	v.value = val
	v.isSet = true
}

func (v NullableClaBudgetLimitRecordListDto) IsSet() bool {
	return v.isSet
}

func (v *NullableClaBudgetLimitRecordListDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableClaBudgetLimitRecordListDto(val *ClaBudgetLimitRecordListDto) *NullableClaBudgetLimitRecordListDto {
	return &NullableClaBudgetLimitRecordListDto{value: val, isSet: true}
}

func (v NullableClaBudgetLimitRecordListDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableClaBudgetLimitRecordListDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
