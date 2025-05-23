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

// checks if the AccountGroupDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &AccountGroupDto{}

// AccountGroupDto struct for AccountGroupDto
type AccountGroupDto struct {
	// The name of the user group.
	GroupName string `json:"groupName"`
	// The UUID of the user group.
	Uuid string `json:"uuid"`
	// The type of the group. `LOCAL`, `SCIM`, `SAML` and `DCS` corresponds to the identity provider from which the group originates. `ALL_USERS` is a special case of `LOCAL` group. It means that group is always assigned to all users in the account.
	Owner string `json:"owner"`
	// The UUID of the Dynatrace account.
	AccountUUID string `json:"accountUUID"`
	// The name of the Dynatrace account.
	AccountName string `json:"accountName"`
	// A short description of the group.
	Description string `json:"description"`
	// The date and time of the group creation in `2021-05-01T15:11:00Z` format.
	CreatedAt string `json:"createdAt"`
	// The date and time of the most recent modification to the group in `2021-05-01T15:11:00Z` format.
	UpdatedAt            string `json:"updatedAt"`
	AdditionalProperties map[string]interface{}
}

type _AccountGroupDto AccountGroupDto

// NewAccountGroupDto instantiates a new AccountGroupDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewAccountGroupDto(groupName string, uuid string, owner string, accountUUID string, accountName string, description string, createdAt string, updatedAt string) *AccountGroupDto {
	this := AccountGroupDto{}
	this.GroupName = groupName
	this.Uuid = uuid
	this.Owner = owner
	this.AccountUUID = accountUUID
	this.AccountName = accountName
	this.Description = description
	this.CreatedAt = createdAt
	this.UpdatedAt = updatedAt
	return &this
}

// NewAccountGroupDtoWithDefaults instantiates a new AccountGroupDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewAccountGroupDtoWithDefaults() *AccountGroupDto {
	this := AccountGroupDto{}
	return &this
}

// GetGroupName returns the GroupName field value
func (o *AccountGroupDto) GetGroupName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.GroupName
}

// GetGroupNameOk returns a tuple with the GroupName field value
// and a boolean to check if the value has been set.
func (o *AccountGroupDto) GetGroupNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.GroupName, true
}

// SetGroupName sets field value
func (o *AccountGroupDto) SetGroupName(v string) {
	o.GroupName = v
}

// GetUuid returns the Uuid field value
func (o *AccountGroupDto) GetUuid() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Uuid
}

// GetUuidOk returns a tuple with the Uuid field value
// and a boolean to check if the value has been set.
func (o *AccountGroupDto) GetUuidOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Uuid, true
}

// SetUuid sets field value
func (o *AccountGroupDto) SetUuid(v string) {
	o.Uuid = v
}

// GetOwner returns the Owner field value
func (o *AccountGroupDto) GetOwner() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Owner
}

// GetOwnerOk returns a tuple with the Owner field value
// and a boolean to check if the value has been set.
func (o *AccountGroupDto) GetOwnerOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Owner, true
}

// SetOwner sets field value
func (o *AccountGroupDto) SetOwner(v string) {
	o.Owner = v
}

// GetAccountUUID returns the AccountUUID field value
func (o *AccountGroupDto) GetAccountUUID() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.AccountUUID
}

// GetAccountUUIDOk returns a tuple with the AccountUUID field value
// and a boolean to check if the value has been set.
func (o *AccountGroupDto) GetAccountUUIDOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.AccountUUID, true
}

// SetAccountUUID sets field value
func (o *AccountGroupDto) SetAccountUUID(v string) {
	o.AccountUUID = v
}

// GetAccountName returns the AccountName field value
func (o *AccountGroupDto) GetAccountName() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.AccountName
}

// GetAccountNameOk returns a tuple with the AccountName field value
// and a boolean to check if the value has been set.
func (o *AccountGroupDto) GetAccountNameOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.AccountName, true
}

// SetAccountName sets field value
func (o *AccountGroupDto) SetAccountName(v string) {
	o.AccountName = v
}

// GetDescription returns the Description field value
func (o *AccountGroupDto) GetDescription() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Description
}

// GetDescriptionOk returns a tuple with the Description field value
// and a boolean to check if the value has been set.
func (o *AccountGroupDto) GetDescriptionOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Description, true
}

// SetDescription sets field value
func (o *AccountGroupDto) SetDescription(v string) {
	o.Description = v
}

// GetCreatedAt returns the CreatedAt field value
func (o *AccountGroupDto) GetCreatedAt() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value
// and a boolean to check if the value has been set.
func (o *AccountGroupDto) GetCreatedAtOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.CreatedAt, true
}

// SetCreatedAt sets field value
func (o *AccountGroupDto) SetCreatedAt(v string) {
	o.CreatedAt = v
}

// GetUpdatedAt returns the UpdatedAt field value
func (o *AccountGroupDto) GetUpdatedAt() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.UpdatedAt
}

// GetUpdatedAtOk returns a tuple with the UpdatedAt field value
// and a boolean to check if the value has been set.
func (o *AccountGroupDto) GetUpdatedAtOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.UpdatedAt, true
}

// SetUpdatedAt sets field value
func (o *AccountGroupDto) SetUpdatedAt(v string) {
	o.UpdatedAt = v
}

func (o AccountGroupDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o AccountGroupDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["groupName"] = o.GroupName
	toSerialize["uuid"] = o.Uuid
	toSerialize["owner"] = o.Owner
	toSerialize["accountUUID"] = o.AccountUUID
	toSerialize["accountName"] = o.AccountName
	toSerialize["description"] = o.Description
	toSerialize["createdAt"] = o.CreatedAt
	toSerialize["updatedAt"] = o.UpdatedAt

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *AccountGroupDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"groupName",
		"uuid",
		"owner",
		"accountUUID",
		"accountName",
		"description",
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

	varAccountGroupDto := _AccountGroupDto{}

	err = json.Unmarshal(data, &varAccountGroupDto)

	if err != nil {
		return err
	}

	*o = AccountGroupDto(varAccountGroupDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "groupName")
		delete(additionalProperties, "uuid")
		delete(additionalProperties, "owner")
		delete(additionalProperties, "accountUUID")
		delete(additionalProperties, "accountName")
		delete(additionalProperties, "description")
		delete(additionalProperties, "createdAt")
		delete(additionalProperties, "updatedAt")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableAccountGroupDto struct {
	value *AccountGroupDto
	isSet bool
}

func (v NullableAccountGroupDto) Get() *AccountGroupDto {
	return v.value
}

func (v *NullableAccountGroupDto) Set(val *AccountGroupDto) {
	v.value = val
	v.isSet = true
}

func (v NullableAccountGroupDto) IsSet() bool {
	return v.isSet
}

func (v *NullableAccountGroupDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableAccountGroupDto(val *AccountGroupDto) *NullableAccountGroupDto {
	return &NullableAccountGroupDto{value: val, isSet: true}
}

func (v NullableAccountGroupDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableAccountGroupDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
