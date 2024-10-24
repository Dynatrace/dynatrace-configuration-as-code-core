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

// checks if the ExternalServiceUserDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ExternalServiceUserDto{}

// ExternalServiceUserDto struct for ExternalServiceUserDto
type ExternalServiceUserDto struct {
	// UUID of service user
	Uid string `json:"uid"`
	// Email of service user
	Email string `json:"email"`
	// Name of service user
	Name string `json:"name"`
	// Surname of service user
	Surname string `json:"surname"`
	// The description of the service user
	Description string `json:"description"`
	// The date and time when the user was created in `2021-05-01T15:11:00Z` format.
	CreatedAt            string `json:"createdAt"`
	AdditionalProperties map[string]interface{}
}

type _ExternalServiceUserDto ExternalServiceUserDto

// NewExternalServiceUserDto instantiates a new ExternalServiceUserDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewExternalServiceUserDto(uid string, email string, name string, surname string, description string, createdAt string) *ExternalServiceUserDto {
	this := ExternalServiceUserDto{}
	this.Uid = uid
	this.Email = email
	this.Name = name
	this.Surname = surname
	this.Description = description
	this.CreatedAt = createdAt
	return &this
}

// NewExternalServiceUserDtoWithDefaults instantiates a new ExternalServiceUserDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewExternalServiceUserDtoWithDefaults() *ExternalServiceUserDto {
	this := ExternalServiceUserDto{}
	return &this
}

// GetUid returns the Uid field value
func (o *ExternalServiceUserDto) GetUid() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Uid
}

// GetUidOk returns a tuple with the Uid field value
// and a boolean to check if the value has been set.
func (o *ExternalServiceUserDto) GetUidOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Uid, true
}

// SetUid sets field value
func (o *ExternalServiceUserDto) SetUid(v string) {
	o.Uid = v
}

// GetEmail returns the Email field value
func (o *ExternalServiceUserDto) GetEmail() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Email
}

// GetEmailOk returns a tuple with the Email field value
// and a boolean to check if the value has been set.
func (o *ExternalServiceUserDto) GetEmailOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Email, true
}

// SetEmail sets field value
func (o *ExternalServiceUserDto) SetEmail(v string) {
	o.Email = v
}

// GetName returns the Name field value
func (o *ExternalServiceUserDto) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *ExternalServiceUserDto) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *ExternalServiceUserDto) SetName(v string) {
	o.Name = v
}

// GetSurname returns the Surname field value
func (o *ExternalServiceUserDto) GetSurname() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Surname
}

// GetSurnameOk returns a tuple with the Surname field value
// and a boolean to check if the value has been set.
func (o *ExternalServiceUserDto) GetSurnameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Surname, true
}

// SetSurname sets field value
func (o *ExternalServiceUserDto) SetSurname(v string) {
	o.Surname = v
}

// GetDescription returns the Description field value
func (o *ExternalServiceUserDto) GetDescription() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Description
}

// GetDescriptionOk returns a tuple with the Description field value
// and a boolean to check if the value has been set.
func (o *ExternalServiceUserDto) GetDescriptionOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Description, true
}

// SetDescription sets field value
func (o *ExternalServiceUserDto) SetDescription(v string) {
	o.Description = v
}

// GetCreatedAt returns the CreatedAt field value
func (o *ExternalServiceUserDto) GetCreatedAt() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value
// and a boolean to check if the value has been set.
func (o *ExternalServiceUserDto) GetCreatedAtOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CreatedAt, true
}

// SetCreatedAt sets field value
func (o *ExternalServiceUserDto) SetCreatedAt(v string) {
	o.CreatedAt = v
}

func (o ExternalServiceUserDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ExternalServiceUserDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["uid"] = o.Uid
	toSerialize["email"] = o.Email
	toSerialize["name"] = o.Name
	toSerialize["surname"] = o.Surname
	toSerialize["description"] = o.Description
	toSerialize["createdAt"] = o.CreatedAt

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *ExternalServiceUserDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"uid",
		"email",
		"name",
		"surname",
		"description",
		"createdAt",
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

	varExternalServiceUserDto := _ExternalServiceUserDto{}

	err = json.Unmarshal(data, &varExternalServiceUserDto)

	if err != nil {
		return err
	}

	*o = ExternalServiceUserDto(varExternalServiceUserDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "uid")
		delete(additionalProperties, "email")
		delete(additionalProperties, "name")
		delete(additionalProperties, "surname")
		delete(additionalProperties, "description")
		delete(additionalProperties, "createdAt")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableExternalServiceUserDto struct {
	value *ExternalServiceUserDto
	isSet bool
}

func (v NullableExternalServiceUserDto) Get() *ExternalServiceUserDto {
	return v.value
}

func (v *NullableExternalServiceUserDto) Set(val *ExternalServiceUserDto) {
	v.value = val
	v.isSet = true
}

func (v NullableExternalServiceUserDto) IsSet() bool {
	return v.isSet
}

func (v *NullableExternalServiceUserDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableExternalServiceUserDto(val *ExternalServiceUserDto) *NullableExternalServiceUserDto {
	return &NullableExternalServiceUserDto{value: val, isSet: true}
}

func (v NullableExternalServiceUserDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableExternalServiceUserDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
