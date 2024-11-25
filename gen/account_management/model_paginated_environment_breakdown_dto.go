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

// checks if the PaginatedEnvironmentBreakdownDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &PaginatedEnvironmentBreakdownDto{}

// PaginatedEnvironmentBreakdownDto struct for PaginatedEnvironmentBreakdownDto
type PaginatedEnvironmentBreakdownDto struct {
	// Identifier of the environment
	EnvironmentId string `json:"environmentId"`
	// Field used to generate the breakdown. Can be `COSTCENTER` or `PRODUCT`
	Field string `json:"field"`
	// List of individual breakdown entries.
	Records []string `json:"records"`
	// Key to request the next page.
	NextPageKey          string `json:"nextPageKey"`
	AdditionalProperties map[string]interface{}
}

type _PaginatedEnvironmentBreakdownDto PaginatedEnvironmentBreakdownDto

// NewPaginatedEnvironmentBreakdownDto instantiates a new PaginatedEnvironmentBreakdownDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewPaginatedEnvironmentBreakdownDto(environmentId string, field string, records []string, nextPageKey string) *PaginatedEnvironmentBreakdownDto {
	this := PaginatedEnvironmentBreakdownDto{}
	this.EnvironmentId = environmentId
	this.Field = field
	this.Records = records
	this.NextPageKey = nextPageKey
	return &this
}

// NewPaginatedEnvironmentBreakdownDtoWithDefaults instantiates a new PaginatedEnvironmentBreakdownDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewPaginatedEnvironmentBreakdownDtoWithDefaults() *PaginatedEnvironmentBreakdownDto {
	this := PaginatedEnvironmentBreakdownDto{}
	return &this
}

// GetEnvironmentId returns the EnvironmentId field value
func (o *PaginatedEnvironmentBreakdownDto) GetEnvironmentId() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.EnvironmentId
}

// GetEnvironmentIdOk returns a tuple with the EnvironmentId field value
// and a boolean to check if the value has been set.
func (o *PaginatedEnvironmentBreakdownDto) GetEnvironmentIdOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.EnvironmentId, true
}

// SetEnvironmentId sets field value
func (o *PaginatedEnvironmentBreakdownDto) SetEnvironmentId(v string) {
	o.EnvironmentId = v
}

// GetField returns the Field field value
func (o *PaginatedEnvironmentBreakdownDto) GetField() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Field
}

// GetFieldOk returns a tuple with the Field field value
// and a boolean to check if the value has been set.
func (o *PaginatedEnvironmentBreakdownDto) GetFieldOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Field, true
}

// SetField sets field value
func (o *PaginatedEnvironmentBreakdownDto) SetField(v string) {
	o.Field = v
}

// GetRecords returns the Records field value
func (o *PaginatedEnvironmentBreakdownDto) GetRecords() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.Records
}

// GetRecordsOk returns a tuple with the Records field value
// and a boolean to check if the value has been set.
func (o *PaginatedEnvironmentBreakdownDto) GetRecordsOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.Records, true
}

// SetRecords sets field value
func (o *PaginatedEnvironmentBreakdownDto) SetRecords(v []string) {
	o.Records = v
}

// GetNextPageKey returns the NextPageKey field value
func (o *PaginatedEnvironmentBreakdownDto) GetNextPageKey() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.NextPageKey
}

// GetNextPageKeyOk returns a tuple with the NextPageKey field value
// and a boolean to check if the value has been set.
func (o *PaginatedEnvironmentBreakdownDto) GetNextPageKeyOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.NextPageKey, true
}

// SetNextPageKey sets field value
func (o *PaginatedEnvironmentBreakdownDto) SetNextPageKey(v string) {
	o.NextPageKey = v
}

func (o PaginatedEnvironmentBreakdownDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o PaginatedEnvironmentBreakdownDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["environmentId"] = o.EnvironmentId
	toSerialize["field"] = o.Field
	toSerialize["records"] = o.Records
	toSerialize["nextPageKey"] = o.NextPageKey

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *PaginatedEnvironmentBreakdownDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"environmentId",
		"field",
		"records",
		"nextPageKey",
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

	varPaginatedEnvironmentBreakdownDto := _PaginatedEnvironmentBreakdownDto{}

	err = json.Unmarshal(data, &varPaginatedEnvironmentBreakdownDto)

	if err != nil {
		return err
	}

	*o = PaginatedEnvironmentBreakdownDto(varPaginatedEnvironmentBreakdownDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "environmentId")
		delete(additionalProperties, "field")
		delete(additionalProperties, "records")
		delete(additionalProperties, "nextPageKey")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullablePaginatedEnvironmentBreakdownDto struct {
	value *PaginatedEnvironmentBreakdownDto
	isSet bool
}

func (v NullablePaginatedEnvironmentBreakdownDto) Get() *PaginatedEnvironmentBreakdownDto {
	return v.value
}

func (v *NullablePaginatedEnvironmentBreakdownDto) Set(val *PaginatedEnvironmentBreakdownDto) {
	v.value = val
	v.isSet = true
}

func (v NullablePaginatedEnvironmentBreakdownDto) IsSet() bool {
	return v.isSet
}

func (v *NullablePaginatedEnvironmentBreakdownDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullablePaginatedEnvironmentBreakdownDto(val *PaginatedEnvironmentBreakdownDto) *NullablePaginatedEnvironmentBreakdownDto {
	return &NullablePaginatedEnvironmentBreakdownDto{value: val, isSet: true}
}

func (v NullablePaginatedEnvironmentBreakdownDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullablePaginatedEnvironmentBreakdownDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}