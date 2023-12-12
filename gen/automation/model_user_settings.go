/*
Automation

Automation API allows working with workflows and various trigger options.

API version: 1.464.0
*/

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package automation

import (
	"encoding/json"
)

// checks if the UserSettings type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &UserSettings{}

// UserSettings struct for UserSettings
type UserSettings struct {
	Groups []string `json:"groups"`
}

// NewUserSettings instantiates a new UserSettings object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewUserSettings(groups []string) *UserSettings {
	this := UserSettings{}
	this.Groups = groups
	return &this
}

// NewUserSettingsWithDefaults instantiates a new UserSettings object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewUserSettingsWithDefaults() *UserSettings {
	this := UserSettings{}
	return &this
}

// GetGroups returns the Groups field value
func (o *UserSettings) GetGroups() []string {
	if o == nil {
		var ret []string
		return ret
	}

	return o.Groups
}

// GetGroupsOk returns a tuple with the Groups field value
// and a boolean to check if the value has been set.
func (o *UserSettings) GetGroupsOk() ([]string, bool) {
	if o == nil {
		return nil, false
	}
	return o.Groups, true
}

// SetGroups sets field value
func (o *UserSettings) SetGroups(v []string) {
	o.Groups = v
}

func (o UserSettings) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o UserSettings) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["groups"] = o.Groups
	return toSerialize, nil
}

type NullableUserSettings struct {
	value *UserSettings
	isSet bool
}

func (v NullableUserSettings) Get() *UserSettings {
	return v.value
}

func (v *NullableUserSettings) Set(val *UserSettings) {
	v.value = val
	v.isSet = true
}

func (v NullableUserSettings) IsSet() bool {
	return v.isSet
}

func (v *NullableUserSettings) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableUserSettings(val *UserSettings) *NullableUserSettings {
	return &NullableUserSettings{value: val, isSet: true}
}

func (v NullableUserSettings) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableUserSettings) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
