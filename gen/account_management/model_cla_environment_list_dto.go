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

// checks if the ClaEnvironmentListDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ClaEnvironmentListDto{}

// ClaEnvironmentListDto struct for ClaEnvironmentListDto
type ClaEnvironmentListDto struct {
	// The number of entries in the list.
	TotalCount float32 `json:"totalCount"`
	// A list of environments of the account.
	Records              []ClaEnvironmentDto `json:"records"`
	AdditionalProperties map[string]interface{}
}

type _ClaEnvironmentListDto ClaEnvironmentListDto

// NewClaEnvironmentListDto instantiates a new ClaEnvironmentListDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewClaEnvironmentListDto(totalCount float32, records []ClaEnvironmentDto) *ClaEnvironmentListDto {
	this := ClaEnvironmentListDto{}
	this.TotalCount = totalCount
	this.Records = records
	return &this
}

// NewClaEnvironmentListDtoWithDefaults instantiates a new ClaEnvironmentListDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewClaEnvironmentListDtoWithDefaults() *ClaEnvironmentListDto {
	this := ClaEnvironmentListDto{}
	return &this
}

// GetTotalCount returns the TotalCount field value
func (o *ClaEnvironmentListDto) GetTotalCount() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.TotalCount
}

// GetTotalCountOk returns a tuple with the TotalCount field value
// and a boolean to check if the value has been set.
func (o *ClaEnvironmentListDto) GetTotalCountOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TotalCount, true
}

// SetTotalCount sets field value
func (o *ClaEnvironmentListDto) SetTotalCount(v float32) {
	o.TotalCount = v
}

// GetRecords returns the Records field value
func (o *ClaEnvironmentListDto) GetRecords() []ClaEnvironmentDto {
	if o == nil {
		var ret []ClaEnvironmentDto
		return ret
	}

	return o.Records
}

// GetRecordsOk returns a tuple with the Records field value
// and a boolean to check if the value has been set.
func (o *ClaEnvironmentListDto) GetRecordsOk() ([]ClaEnvironmentDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Records, true
}

// SetRecords sets field value
func (o *ClaEnvironmentListDto) SetRecords(v []ClaEnvironmentDto) {
	o.Records = v
}

func (o ClaEnvironmentListDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ClaEnvironmentListDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["totalCount"] = o.TotalCount
	toSerialize["records"] = o.Records

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *ClaEnvironmentListDto) UnmarshalJSON(data []byte) (err error) {
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

	varClaEnvironmentListDto := _ClaEnvironmentListDto{}

	err = json.Unmarshal(data, &varClaEnvironmentListDto)

	if err != nil {
		return err
	}

	*o = ClaEnvironmentListDto(varClaEnvironmentListDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "totalCount")
		delete(additionalProperties, "records")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableClaEnvironmentListDto struct {
	value *ClaEnvironmentListDto
	isSet bool
}

func (v NullableClaEnvironmentListDto) Get() *ClaEnvironmentListDto {
	return v.value
}

func (v *NullableClaEnvironmentListDto) Set(val *ClaEnvironmentListDto) {
	v.value = val
	v.isSet = true
}

func (v NullableClaEnvironmentListDto) IsSet() bool {
	return v.isSet
}

func (v *NullableClaEnvironmentListDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableClaEnvironmentListDto(val *ClaEnvironmentListDto) *NullableClaEnvironmentListDto {
	return &NullableClaEnvironmentListDto{value: val, isSet: true}
}

func (v NullableClaEnvironmentListDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableClaEnvironmentListDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
