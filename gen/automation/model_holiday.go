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

// checks if the Holiday type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &Holiday{}

// Holiday struct for Holiday
type Holiday struct {
	Title string `json:"title"`
	Date  string `json:"date"`
}

// NewHoliday instantiates a new Holiday object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewHoliday(title string, date string) *Holiday {
	this := Holiday{}
	this.Title = title
	this.Date = date
	return &this
}

// NewHolidayWithDefaults instantiates a new Holiday object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewHolidayWithDefaults() *Holiday {
	this := Holiday{}
	return &this
}

// GetTitle returns the Title field value
func (o *Holiday) GetTitle() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Title
}

// GetTitleOk returns a tuple with the Title field value
// and a boolean to check if the value has been set.
func (o *Holiday) GetTitleOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Title, true
}

// SetTitle sets field value
func (o *Holiday) SetTitle(v string) {
	o.Title = v
}

// GetDate returns the Date field value
func (o *Holiday) GetDate() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Date
}

// GetDateOk returns a tuple with the Date field value
// and a boolean to check if the value has been set.
func (o *Holiday) GetDateOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Date, true
}

// SetDate sets field value
func (o *Holiday) SetDate(v string) {
	o.Date = v
}

func (o Holiday) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o Holiday) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	toSerialize["title"] = o.Title
	toSerialize["date"] = o.Date
	return toSerialize, nil
}

type NullableHoliday struct {
	value *Holiday
	isSet bool
}

func (v NullableHoliday) Get() *Holiday {
	return v.value
}

func (v *NullableHoliday) Set(val *Holiday) {
	v.value = val
	v.isSet = true
}

func (v NullableHoliday) IsSet() bool {
	return v.isSet
}

func (v *NullableHoliday) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableHoliday(val *Holiday) *NullableHoliday {
	return &NullableHoliday{value: val, isSet: true}
}

func (v NullableHoliday) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableHoliday) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
