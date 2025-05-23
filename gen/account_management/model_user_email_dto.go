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

// checks if the UserEmailDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UserEmailDto{}

// UserEmailDto struct for UserEmailDto
type UserEmailDto struct {
	// The email address of the user.
	Email                string `json:"email"`
	AdditionalProperties map[string]interface{}
}

type _UserEmailDto UserEmailDto

// NewUserEmailDto instantiates a new UserEmailDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUserEmailDto(email string) *UserEmailDto {
	this := UserEmailDto{}
	this.Email = email
	return &this
}

// NewUserEmailDtoWithDefaults instantiates a new UserEmailDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUserEmailDtoWithDefaults() *UserEmailDto {
	this := UserEmailDto{}
	return &this
}

// GetEmail returns the Email field value
func (o *UserEmailDto) GetEmail() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Email
}

// GetEmailOk returns a tuple with the Email field value
// and a boolean to check if the value has been set.
func (o *UserEmailDto) GetEmailOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Email, true
}

// SetEmail sets field value
func (o *UserEmailDto) SetEmail(v string) {
	o.Email = v
}

func (o UserEmailDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UserEmailDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["email"] = o.Email

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *UserEmailDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"email",
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

	varUserEmailDto := _UserEmailDto{}

	err = json.Unmarshal(data, &varUserEmailDto)

	if err != nil {
		return err
	}

	*o = UserEmailDto(varUserEmailDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "email")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableUserEmailDto struct {
	value *UserEmailDto
	isSet bool
}

func (v NullableUserEmailDto) Get() *UserEmailDto {
	return v.value
}

func (v *NullableUserEmailDto) Set(val *UserEmailDto) {
	v.value = val
	v.isSet = true
}

func (v NullableUserEmailDto) IsSet() bool {
	return v.isSet
}

func (v *NullableUserEmailDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUserEmailDto(val *UserEmailDto) *NullableUserEmailDto {
	return &NullableUserEmailDto{value: val, isSet: true}
}

func (v NullableUserEmailDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUserEmailDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
