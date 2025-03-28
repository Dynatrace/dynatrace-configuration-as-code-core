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

// checks if the EnvironmentResourceDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &EnvironmentResourceDto{}

// EnvironmentResourceDto struct for EnvironmentResourceDto
type EnvironmentResourceDto struct {
	// A list of environments in the account.
	TenantResources []TenantResourceDto `json:"tenantResources"`
	// A list of management zones in the account.
	ManagementZoneResources []ManagementZoneResourceDto `json:"managementZoneResources"`
	AdditionalProperties    map[string]interface{}
}

type _EnvironmentResourceDto EnvironmentResourceDto

// NewEnvironmentResourceDto instantiates a new EnvironmentResourceDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewEnvironmentResourceDto(tenantResources []TenantResourceDto, managementZoneResources []ManagementZoneResourceDto) *EnvironmentResourceDto {
	this := EnvironmentResourceDto{}
	this.TenantResources = tenantResources
	this.ManagementZoneResources = managementZoneResources
	return &this
}

// NewEnvironmentResourceDtoWithDefaults instantiates a new EnvironmentResourceDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewEnvironmentResourceDtoWithDefaults() *EnvironmentResourceDto {
	this := EnvironmentResourceDto{}
	return &this
}

// GetTenantResources returns the TenantResources field value
func (o *EnvironmentResourceDto) GetTenantResources() []TenantResourceDto {
	if o == nil {
		var ret []TenantResourceDto
		return ret
	}

	return o.TenantResources
}

// GetTenantResourcesOk returns a tuple with the TenantResources field value
// and a boolean to check if the value has been set.
func (o *EnvironmentResourceDto) GetTenantResourcesOk() ([]TenantResourceDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.TenantResources, true
}

// SetTenantResources sets field value
func (o *EnvironmentResourceDto) SetTenantResources(v []TenantResourceDto) {
	o.TenantResources = v
}

// GetManagementZoneResources returns the ManagementZoneResources field value
func (o *EnvironmentResourceDto) GetManagementZoneResources() []ManagementZoneResourceDto {
	if o == nil {
		var ret []ManagementZoneResourceDto
		return ret
	}

	return o.ManagementZoneResources
}

// GetManagementZoneResourcesOk returns a tuple with the ManagementZoneResources field value
// and a boolean to check if the value has been set.
func (o *EnvironmentResourceDto) GetManagementZoneResourcesOk() ([]ManagementZoneResourceDto, bool) {
	if o == nil {
		return nil, false
	}
	return o.ManagementZoneResources, true
}

// SetManagementZoneResources sets field value
func (o *EnvironmentResourceDto) SetManagementZoneResources(v []ManagementZoneResourceDto) {
	o.ManagementZoneResources = v
}

func (o EnvironmentResourceDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o EnvironmentResourceDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["tenantResources"] = o.TenantResources
	toSerialize["managementZoneResources"] = o.ManagementZoneResources

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *EnvironmentResourceDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"tenantResources",
		"managementZoneResources",
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

	varEnvironmentResourceDto := _EnvironmentResourceDto{}

	err = json.Unmarshal(data, &varEnvironmentResourceDto)

	if err != nil {
		return err
	}

	*o = EnvironmentResourceDto(varEnvironmentResourceDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "tenantResources")
		delete(additionalProperties, "managementZoneResources")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableEnvironmentResourceDto struct {
	value *EnvironmentResourceDto
	isSet bool
}

func (v NullableEnvironmentResourceDto) Get() *EnvironmentResourceDto {
	return v.value
}

func (v *NullableEnvironmentResourceDto) Set(val *EnvironmentResourceDto) {
	v.value = val
	v.isSet = true
}

func (v NullableEnvironmentResourceDto) IsSet() bool {
	return v.isSet
}

func (v *NullableEnvironmentResourceDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableEnvironmentResourceDto(val *EnvironmentResourceDto) *NullableEnvironmentResourceDto {
	return &NullableEnvironmentResourceDto{value: val, isSet: true}
}

func (v NullableEnvironmentResourceDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableEnvironmentResourceDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
