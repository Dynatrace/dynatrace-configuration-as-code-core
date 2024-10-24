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

// checks if the PaginatedFieldValueDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PaginatedFieldValueDto{}

// PaginatedFieldValueDto struct for PaginatedFieldValueDto
type PaginatedFieldValueDto struct {
	// The records on the current page.
	Records []FieldValueDto `json:"records"`
	// Indicates if there is another page to load.
	HasNextPage          bool `json:"hasNextPage"`
	AdditionalProperties map[string]interface{}
}

type _PaginatedFieldValueDto PaginatedFieldValueDto

// NewPaginatedFieldValueDto instantiates a new PaginatedFieldValueDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPaginatedFieldValueDto(records []FieldValueDto, hasNextPage bool) *PaginatedFieldValueDto {
	this := PaginatedFieldValueDto{}
	this.Records = records
	this.HasNextPage = hasNextPage
	return &this
}

// NewPaginatedFieldValueDtoWithDefaults instantiates a new PaginatedFieldValueDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPaginatedFieldValueDtoWithDefaults() *PaginatedFieldValueDto {
	this := PaginatedFieldValueDto{}
	return &this
}

// GetRecords returns the Records field value
func (o *PaginatedFieldValueDto) GetRecords() []FieldValueDto {
	if o == nil {
		var ret []FieldValueDto
		return ret
	}

	return o.Records
}

// GetRecordsOk returns a tuple with the Records field value
// and a boolean to check if the value has been set.
func (o *PaginatedFieldValueDto) GetRecordsOk() ([]FieldValueDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Records, true
}

// SetRecords sets field value
func (o *PaginatedFieldValueDto) SetRecords(v []FieldValueDto) {
	o.Records = v
}

// GetHasNextPage returns the HasNextPage field value
func (o *PaginatedFieldValueDto) GetHasNextPage() bool {
	if o == nil {
		var ret bool
		return ret
	}

	return o.HasNextPage
}

// GetHasNextPageOk returns a tuple with the HasNextPage field value
// and a boolean to check if the value has been set.
func (o *PaginatedFieldValueDto) GetHasNextPageOk() (*bool, bool) {
	if o == nil {
		return nil, false
	}
	return &o.HasNextPage, true
}

// SetHasNextPage sets field value
func (o *PaginatedFieldValueDto) SetHasNextPage(v bool) {
	o.HasNextPage = v
}

func (o PaginatedFieldValueDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PaginatedFieldValueDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["records"] = o.Records
	toSerialize["hasNextPage"] = o.HasNextPage

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *PaginatedFieldValueDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"records",
		"hasNextPage",
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

	varPaginatedFieldValueDto := _PaginatedFieldValueDto{}

	err = json.Unmarshal(data, &varPaginatedFieldValueDto)

	if err != nil {
		return err
	}

	*o = PaginatedFieldValueDto(varPaginatedFieldValueDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "records")
		delete(additionalProperties, "hasNextPage")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullablePaginatedFieldValueDto struct {
	value *PaginatedFieldValueDto
	isSet bool
}

func (v NullablePaginatedFieldValueDto) Get() *PaginatedFieldValueDto {
	return v.value
}

func (v *NullablePaginatedFieldValueDto) Set(val *PaginatedFieldValueDto) {
	v.value = val
	v.isSet = true
}

func (v NullablePaginatedFieldValueDto) IsSet() bool {
	return v.isSet
}

func (v *NullablePaginatedFieldValueDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePaginatedFieldValueDto(val *PaginatedFieldValueDto) *NullablePaginatedFieldValueDto {
	return &NullablePaginatedFieldValueDto{value: val, isSet: true}
}

func (v NullablePaginatedFieldValueDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePaginatedFieldValueDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
