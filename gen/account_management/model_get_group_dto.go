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

// checks if the GetGroupDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &GetGroupDto{}

// GetGroupDto struct for GetGroupDto
type GetGroupDto struct {
	// The UUID of the user group.
	Uuid *string `json:"uuid,omitempty"`
	// The name of the user group.
	Name string `json:"name"`
	// A short description of the user group.
	Description *string `json:"description,omitempty"`
	// A list of values associating this group with the corresponding claim from an identity provider.
	FederatedAttributeValues []string `json:"federatedAttributeValues,omitempty"`
	// The type of the group. `LOCAL`, `SCIM`, `SAML` and `DCS` corresponds to the identity provider from which the group originates. `ALL_USERS` is a special case of `LOCAL` group. It means that group is always assigned to all users in the account.
	Owner string `json:"owner"`
	// The date and time of the group creation in `2021-05-01T15:11:00Z` format.
	CreatedAt string `json:"createdAt"`
	// The date and time of the most recent group modification in `2021-05-01T15:11:00Z` format.
	UpdatedAt            string `json:"updatedAt"`
	AdditionalProperties map[string]interface{}
}

type _GetGroupDto GetGroupDto

// NewGetGroupDto instantiates a new GetGroupDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewGetGroupDto(name string, owner string, createdAt string, updatedAt string) *GetGroupDto {
	this := GetGroupDto{}
	this.Name = name
	this.Owner = owner
	this.CreatedAt = createdAt
	this.UpdatedAt = updatedAt
	return &this
}

// NewGetGroupDtoWithDefaults instantiates a new GetGroupDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewGetGroupDtoWithDefaults() *GetGroupDto {
	this := GetGroupDto{}
	return &this
}

// GetUuid returns the Uuid field value if set, zero value otherwise.
func (o *GetGroupDto) GetUuid() string {
	if o == nil || IsNil(o.Uuid) {
		var ret string
		return ret
	}
	return *o.Uuid
}

// GetUuidOk returns a tuple with the Uuid field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetGroupDto) GetUuidOk() (*string, bool) {
	if o == nil || IsNil(o.Uuid) {
		return nil, false
	}
	return o.Uuid, true
}

// HasUuid returns a boolean if a field has been set.
func (o *GetGroupDto) HasUuid() bool {
	if o != nil && !IsNil(o.Uuid) {
		return true
	}

	return false
}

// SetUuid gets a reference to the given string and assigns it to the Uuid field.
func (o *GetGroupDto) SetUuid(v string) {
	o.Uuid = &v
}

// GetName returns the Name field value
func (o *GetGroupDto) GetName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Name
}

// GetNameOk returns a tuple with the Name field value
// and a boolean to check if the value has been set.
func (o *GetGroupDto) GetNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Name, true
}

// SetName sets field value
func (o *GetGroupDto) SetName(v string) {
	o.Name = v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *GetGroupDto) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetGroupDto) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *GetGroupDto) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *GetGroupDto) SetDescription(v string) {
	o.Description = &v
}

// GetFederatedAttributeValues returns the FederatedAttributeValues field value if set, zero value otherwise.
func (o *GetGroupDto) GetFederatedAttributeValues() []string {
	if o == nil || IsNil(o.FederatedAttributeValues) {
		var ret []string
		return ret
	}
	return o.FederatedAttributeValues
}

// GetFederatedAttributeValuesOk returns a tuple with the FederatedAttributeValues field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *GetGroupDto) GetFederatedAttributeValuesOk() ([]string, bool) {
	if o == nil || IsNil(o.FederatedAttributeValues) {
		return nil, false
	}
	return o.FederatedAttributeValues, true
}

// HasFederatedAttributeValues returns a boolean if a field has been set.
func (o *GetGroupDto) HasFederatedAttributeValues() bool {
	if o != nil && !IsNil(o.FederatedAttributeValues) {
		return true
	}

	return false
}

// SetFederatedAttributeValues gets a reference to the given []string and assigns it to the FederatedAttributeValues field.
func (o *GetGroupDto) SetFederatedAttributeValues(v []string) {
	o.FederatedAttributeValues = v
}

// GetOwner returns the Owner field value
func (o *GetGroupDto) GetOwner() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Owner
}

// GetOwnerOk returns a tuple with the Owner field value
// and a boolean to check if the value has been set.
func (o *GetGroupDto) GetOwnerOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Owner, true
}

// SetOwner sets field value
func (o *GetGroupDto) SetOwner(v string) {
	o.Owner = v
}

// GetCreatedAt returns the CreatedAt field value
func (o *GetGroupDto) GetCreatedAt() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value
// and a boolean to check if the value has been set.
func (o *GetGroupDto) GetCreatedAtOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CreatedAt, true
}

// SetCreatedAt sets field value
func (o *GetGroupDto) SetCreatedAt(v string) {
	o.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt field value
func (o *GetGroupDto) GetUpdatedAt() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.UpdatedAt
}

// GetUpdatedAtOk returns a tuple with the UpdatedAt field value
// and a boolean to check if the value has been set.
func (o *GetGroupDto) GetUpdatedAtOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UpdatedAt, true
}

// SetUpdatedAt sets field value
func (o *GetGroupDto) SetUpdatedAt(v string) {
	o.UpdatedAt = v
}

func (o GetGroupDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o GetGroupDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Uuid) {
		toSerialize["uuid"] = o.Uuid
	}
	toSerialize["name"] = o.Name
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	if !IsNil(o.FederatedAttributeValues) {
		toSerialize["federatedAttributeValues"] = o.FederatedAttributeValues
	}
	toSerialize["owner"] = o.Owner
	toSerialize["createdAt"] = o.CreatedAt
	toSerialize["updatedAt"] = o.UpdatedAt

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *GetGroupDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"name",
		"owner",
		"createdAt",
		"updatedAt",
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

	varGetGroupDto := _GetGroupDto{}

	err = json.Unmarshal(data, &varGetGroupDto)

	if err != nil {
		return err
	}

	*o = GetGroupDto(varGetGroupDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "uuid")
		delete(additionalProperties, "name")
		delete(additionalProperties, "description")
		delete(additionalProperties, "federatedAttributeValues")
		delete(additionalProperties, "owner")
		delete(additionalProperties, "createdAt")
		delete(additionalProperties, "updatedAt")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableGetGroupDto struct {
	value *GetGroupDto
	isSet bool
}

func (v NullableGetGroupDto) Get() *GetGroupDto {
	return v.value
}

func (v *NullableGetGroupDto) Set(val *GetGroupDto) {
	v.value = val
	v.isSet = true
}

func (v NullableGetGroupDto) IsSet() bool {
	return v.isSet
}

func (v *NullableGetGroupDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableGetGroupDto(val *GetGroupDto) *NullableGetGroupDto {
	return &NullableGetGroupDto{value: val, isSet: true}
}

func (v NullableGetGroupDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableGetGroupDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
