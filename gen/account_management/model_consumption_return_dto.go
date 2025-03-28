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
	"time"
)

// checks if the ConsumptionReturnDto type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &ConsumptionReturnDto{}

// ConsumptionReturnDto struct for ConsumptionReturnDto
type ConsumptionReturnDto struct {
	// The start date and time of the report interval in `2021-05-01T15:11:00Z` format.
	TimeFrameStart time.Time `json:"timeFrameStart"`
	// The end date and time of the report interval in `2021-05-01T15:11:00Z` format.
	TimeFrameEnd time.Time `json:"timeFrameEnd"`
	// The name of the consumed units (for example, `Davis data units`).
	ConsumptionType string `json:"consumptionType"`
	// The quantity that has been deducted from the available unit's pool.
	Quantity             float32 `json:"quantity"`
	AdditionalProperties map[string]interface{}
}

type _ConsumptionReturnDto ConsumptionReturnDto

// NewConsumptionReturnDto instantiates a new ConsumptionReturnDto object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewConsumptionReturnDto(timeFrameStart time.Time, timeFrameEnd time.Time, consumptionType string, quantity float32) *ConsumptionReturnDto {
	this := ConsumptionReturnDto{}
	this.TimeFrameStart = timeFrameStart
	this.TimeFrameEnd = timeFrameEnd
	this.ConsumptionType = consumptionType
	this.Quantity = quantity
	return &this
}

// NewConsumptionReturnDtoWithDefaults instantiates a new ConsumptionReturnDto object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewConsumptionReturnDtoWithDefaults() *ConsumptionReturnDto {
	this := ConsumptionReturnDto{}
	return &this
}

// GetTimeFrameStart returns the TimeFrameStart field value
func (o *ConsumptionReturnDto) GetTimeFrameStart() time.Time {
	if o == nil {
		var ret time.Time
		return ret
	}

	return o.TimeFrameStart
}

// GetTimeFrameStartOk returns a tuple with the TimeFrameStart field value
// and a boolean to check if the value has been set.
func (o *ConsumptionReturnDto) GetTimeFrameStartOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TimeFrameStart, true
}

// SetTimeFrameStart sets field value
func (o *ConsumptionReturnDto) SetTimeFrameStart(v time.Time) {
	o.TimeFrameStart = v
}

// GetTimeFrameEnd returns the TimeFrameEnd field value
func (o *ConsumptionReturnDto) GetTimeFrameEnd() time.Time {
	if o == nil {
		var ret time.Time
		return ret
	}

	return o.TimeFrameEnd
}

// GetTimeFrameEndOk returns a tuple with the TimeFrameEnd field value
// and a boolean to check if the value has been set.
func (o *ConsumptionReturnDto) GetTimeFrameEndOk() (*time.Time, bool) {
	if o == nil {
		return nil, false
	}
	return &o.TimeFrameEnd, true
}

// SetTimeFrameEnd sets field value
func (o *ConsumptionReturnDto) SetTimeFrameEnd(v time.Time) {
	o.TimeFrameEnd = v
}

// GetConsumptionType returns the ConsumptionType field value
func (o *ConsumptionReturnDto) GetConsumptionType() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.ConsumptionType
}

// GetConsumptionTypeOk returns a tuple with the ConsumptionType field value
// and a boolean to check if the value has been set.
func (o *ConsumptionReturnDto) GetConsumptionTypeOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.ConsumptionType, true
}

// SetConsumptionType sets field value
func (o *ConsumptionReturnDto) SetConsumptionType(v string) {
	o.ConsumptionType = v
}

// GetQuantity returns the Quantity field value
func (o *ConsumptionReturnDto) GetQuantity() float32 {
	if o == nil {
		var ret float32
		return ret
	}

	return o.Quantity
}

// GetQuantityOk returns a tuple with the Quantity field value
// and a boolean to check if the value has been set.
func (o *ConsumptionReturnDto) GetQuantityOk() (*float32, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Quantity, true
}

// SetQuantity sets field value
func (o *ConsumptionReturnDto) SetQuantity(v float32) {
	o.Quantity = v
}

func (o ConsumptionReturnDto) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o ConsumptionReturnDto) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["timeFrameStart"] = o.TimeFrameStart
	toSerialize["timeFrameEnd"] = o.TimeFrameEnd
	toSerialize["consumptionType"] = o.ConsumptionType
	toSerialize["quantity"] = o.Quantity

	for key, value := range o.AdditionalProperties {
		toSerialize[key] = value
	}

	return toSerialize, nil
}

func (o *ConsumptionReturnDto) UnmarshalJSON(data []byte) (err error) {
	// This validates that all required properties are included in the JSON object
	// by unmarshalling the object into a generic map with string keys and checking
	// that every required field exists as a key in the generic map.
	requiredProperties := []string{
		"timeFrameStart",
		"timeFrameEnd",
		"consumptionType",
		"quantity",
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

	varConsumptionReturnDto := _ConsumptionReturnDto{}

	err = json.Unmarshal(data, &varConsumptionReturnDto)

	if err != nil {
		return err
	}

	*o = ConsumptionReturnDto(varConsumptionReturnDto)

	additionalProperties := make(map[string]interface{})

	if err = json.Unmarshal(data, &additionalProperties); err == nil {
		delete(additionalProperties, "timeFrameStart")
		delete(additionalProperties, "timeFrameEnd")
		delete(additionalProperties, "consumptionType")
		delete(additionalProperties, "quantity")
		o.AdditionalProperties = additionalProperties
	}

	return err
}

type NullableConsumptionReturnDto struct {
	value *ConsumptionReturnDto
	isSet bool
}

func (v NullableConsumptionReturnDto) Get() *ConsumptionReturnDto {
	return v.value
}

func (v *NullableConsumptionReturnDto) Set(val *ConsumptionReturnDto) {
	v.value = val
	v.isSet = true
}

func (v NullableConsumptionReturnDto) IsSet() bool {
	return v.isSet
}

func (v *NullableConsumptionReturnDto) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableConsumptionReturnDto(val *ConsumptionReturnDto) *NullableConsumptionReturnDto {
	return &NullableConsumptionReturnDto{value: val, isSet: true}
}

func (v NullableConsumptionReturnDto) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableConsumptionReturnDto) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
