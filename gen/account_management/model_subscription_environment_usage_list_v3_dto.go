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

// checks if the SubscriptionEnvironmentUsageListV3Dto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &SubscriptionEnvironmentUsageListV3Dto{}

// SubscriptionEnvironmentUsageListV3Dto struct for SubscriptionEnvironmentUsageListV3Dto
type SubscriptionEnvironmentUsageListV3Dto struct {
	// Subscription usage data
	Data []SubscriptionEnvironmentUsageV3Dto `json:"data"`
	// The time the subscription data was last modified in `2021-05-01T15:11:00Z` format.
	LastModifiedTime string `json:"lastModifiedTime"`
	// The next page key for pagination if next page exists
	NextPageKey          string `json:"nextPageKey"`
	AdditionalProperties map[string]interface{}
}

type _SubscriptionEnvironmentUsageListV3Dto SubscriptionEnvironmentUsageListV3Dto

// NewSubscriptionEnvironmentUsageListV3Dto instantiates a new SubscriptionEnvironmentUsageListV3Dto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewSubscriptionEnvironmentUsageListV3Dto(data []SubscriptionEnvironmentUsageV3Dto, lastModifiedTime string, nextPageKey string) *SubscriptionEnvironmentUsageListV3Dto {
	this := SubscriptionEnvironmentUsageListV3Dto{}
	this.Data = data
	this.LastModifiedTime = lastModifiedTime
	this.NextPageKey = nextPageKey
	return &this
}

// NewSubscriptionEnvironmentUsageListV3DtoWithDefaults instantiates a new SubscriptionEnvironmentUsageListV3Dto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewSubscriptionEnvironmentUsageListV3DtoWithDefaults() *SubscriptionEnvironmentUsageListV3Dto {
	this := SubscriptionEnvironmentUsageListV3Dto{}
	return &this
}

// GetData returns the Data field value
func (o *SubscriptionEnvironmentUsageListV3Dto) GetData() []SubscriptionEnvironmentUsageV3Dto {
	if o == nil {
		var ret []SubscriptionEnvironmentUsageV3Dto
		return ret
	}

	return o.Data
}

// GetDataOk returns a tuple with the Data field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentUsageListV3Dto) GetDataOk() ([]SubscriptionEnvironmentUsageV3Dto, bool) {
	if o == nil {
		return nil, false
	}
	return o.Data, true
}

// SetData sets field value
func (o *SubscriptionEnvironmentUsageListV3Dto) SetData(v []SubscriptionEnvironmentUsageV3Dto) {
	o.Data = v
}

// GetLastModifiedTime returns the LastModifiedTime field value
func (o *SubscriptionEnvironmentUsageListV3Dto) GetLastModifiedTime() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.LastModifiedTime
}

// GetLastModifiedTimeOk returns a tuple with the LastModifiedTime field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentUsageListV3Dto) GetLastModifiedTimeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.LastModifiedTime, true
}

// SetLastModifiedTime sets field value
func (o *SubscriptionEnvironmentUsageListV3Dto) SetLastModifiedTime(v string) {
	o.LastModifiedTime = v
}

// GetNextPageKey returns the NextPageKey field value
func (o *SubscriptionEnvironmentUsageListV3Dto) GetNextPageKey() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.NextPageKey
}

// GetNextPageKeyOk returns a tuple with the NextPageKey field value
// and a boolean to check if the value has been set.
func (o *SubscriptionEnvironmentUsageListV3Dto) GetNextPageKeyOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.NextPageKey, true
}

// SetNextPageKey sets field value
func (o *SubscriptionEnvironmentUsageListV3Dto) SetNextPageKey(v string) {
	o.NextPageKey = v
}

func (o SubscriptionEnvironmentUsageListV3Dto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o SubscriptionEnvironmentUsageListV3Dto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["data"] = o.Data
	toSerialize["lastModifiedTime"] = o.LastModifiedTime
	toSerialize["nextPageKey"] = o.NextPageKey

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *SubscriptionEnvironmentUsageListV3Dto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"data",
		"lastModifiedTime",
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

	varSubscriptionEnvironmentUsageListV3Dto := _SubscriptionEnvironmentUsageListV3Dto{}

	err = json.Unmarshal(data, &varSubscriptionEnvironmentUsageListV3Dto)

	if err != nil {
		return err
	}

	*o = SubscriptionEnvironmentUsageListV3Dto(varSubscriptionEnvironmentUsageListV3Dto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "data")
		delete(additionalProperties, "lastModifiedTime")
		delete(additionalProperties, "nextPageKey")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableSubscriptionEnvironmentUsageListV3Dto struct {
	value *SubscriptionEnvironmentUsageListV3Dto
	isSet bool
}

func (v NullableSubscriptionEnvironmentUsageListV3Dto) Get() *SubscriptionEnvironmentUsageListV3Dto {
	return v.value
}

func (v *NullableSubscriptionEnvironmentUsageListV3Dto) Set(val *SubscriptionEnvironmentUsageListV3Dto) {
	v.value = val
	v.isSet = true
}

func (v NullableSubscriptionEnvironmentUsageListV3Dto) IsSet() bool {
	return v.isSet
}

func (v *NullableSubscriptionEnvironmentUsageListV3Dto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableSubscriptionEnvironmentUsageListV3Dto(val *SubscriptionEnvironmentUsageListV3Dto) *NullableSubscriptionEnvironmentUsageListV3Dto {
	return &NullableSubscriptionEnvironmentUsageListV3Dto{value: val, isSet: true}
}

func (v NullableSubscriptionEnvironmentUsageListV3Dto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableSubscriptionEnvironmentUsageListV3Dto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
