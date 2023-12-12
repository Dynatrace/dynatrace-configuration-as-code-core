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

// checks if the BusinessCalendarRequest type satisfies the MappedNullable interface at compile time
var _ MappedNullable = &BusinessCalendarRequest{}

// BusinessCalendarRequest struct for BusinessCalendarRequest
type BusinessCalendarRequest struct {
	Id          *string   `json:"id,omitempty"`
	Title       string    `json:"title"`
	Weekstart   *int32    `json:"weekstart,omitempty"`
	Weekdays    []string  `json:"weekdays,omitempty"`
	Holidays    []Holiday `json:"holidays,omitempty"`
	ValidFrom   *string   `json:"validFrom,omitempty"`
	ValidTo     *string   `json:"validTo,omitempty"`
	Description *string   `json:"description,omitempty"`
}

// NewBusinessCalendarRequest instantiates a new BusinessCalendarRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewBusinessCalendarRequest(title string) *BusinessCalendarRequest {
	this := BusinessCalendarRequest{}
	this.Title = title
	return &this
}

// NewBusinessCalendarRequestWithDefaults instantiates a new BusinessCalendarRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewBusinessCalendarRequestWithDefaults() *BusinessCalendarRequest {
	this := BusinessCalendarRequest{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *BusinessCalendarRequest) GetId() string {
	if o == nil || IsNil(o.Id) {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BusinessCalendarRequest) GetIdOk() (*string, bool) {
	if o == nil || IsNil(o.Id) {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *BusinessCalendarRequest) HasId() bool {
	if o != nil && !IsNil(o.Id) {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *BusinessCalendarRequest) SetId(v string) {
	o.Id = &v
}

// GetTitle returns the Title field value
func (o *BusinessCalendarRequest) GetTitle() string {
	if o == nil {
		var ret string
		return ret
	}

	return o.Title
}

// GetTitleOk returns a tuple with the Title field value
// and a boolean to check if the value has been set.
func (o *BusinessCalendarRequest) GetTitleOk() (*string, bool) {
	if o == nil {
		return nil, false
	}
	return &o.Title, true
}

// SetTitle sets field value
func (o *BusinessCalendarRequest) SetTitle(v string) {
	o.Title = v
}

// GetWeekstart returns the Weekstart field value if set, zero value otherwise.
func (o *BusinessCalendarRequest) GetWeekstart() int32 {
	if o == nil || IsNil(o.Weekstart) {
		var ret int32
		return ret
	}
	return *o.Weekstart
}

// GetWeekstartOk returns a tuple with the Weekstart field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BusinessCalendarRequest) GetWeekstartOk() (*int32, bool) {
	if o == nil || IsNil(o.Weekstart) {
		return nil, false
	}
	return o.Weekstart, true
}

// HasWeekstart returns a boolean if a field has been set.
func (o *BusinessCalendarRequest) HasWeekstart() bool {
	if o != nil && !IsNil(o.Weekstart) {
		return true
	}

	return false
}

// SetWeekstart gets a reference to the given int32 and assigns it to the Weekstart field.
func (o *BusinessCalendarRequest) SetWeekstart(v int32) {
	o.Weekstart = &v
}

// GetWeekdays returns the Weekdays field value if set, zero value otherwise.
func (o *BusinessCalendarRequest) GetWeekdays() []string {
	if o == nil || IsNil(o.Weekdays) {
		var ret []string
		return ret
	}
	return o.Weekdays
}

// GetWeekdaysOk returns a tuple with the Weekdays field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BusinessCalendarRequest) GetWeekdaysOk() ([]string, bool) {
	if o == nil || IsNil(o.Weekdays) {
		return nil, false
	}
	return o.Weekdays, true
}

// HasWeekdays returns a boolean if a field has been set.
func (o *BusinessCalendarRequest) HasWeekdays() bool {
	if o != nil && !IsNil(o.Weekdays) {
		return true
	}

	return false
}

// SetWeekdays gets a reference to the given []string and assigns it to the Weekdays field.
func (o *BusinessCalendarRequest) SetWeekdays(v []string) {
	o.Weekdays = v
}

// GetHolidays returns the Holidays field value if set, zero value otherwise.
func (o *BusinessCalendarRequest) GetHolidays() []Holiday {
	if o == nil || IsNil(o.Holidays) {
		var ret []Holiday
		return ret
	}
	return o.Holidays
}

// GetHolidaysOk returns a tuple with the Holidays field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BusinessCalendarRequest) GetHolidaysOk() ([]Holiday, bool) {
	if o == nil || IsNil(o.Holidays) {
		return nil, false
	}
	return o.Holidays, true
}

// HasHolidays returns a boolean if a field has been set.
func (o *BusinessCalendarRequest) HasHolidays() bool {
	if o != nil && !IsNil(o.Holidays) {
		return true
	}

	return false
}

// SetHolidays gets a reference to the given []Holiday and assigns it to the Holidays field.
func (o *BusinessCalendarRequest) SetHolidays(v []Holiday) {
	o.Holidays = v
}

// GetValidFrom returns the ValidFrom field value if set, zero value otherwise.
func (o *BusinessCalendarRequest) GetValidFrom() string {
	if o == nil || IsNil(o.ValidFrom) {
		var ret string
		return ret
	}
	return *o.ValidFrom
}

// GetValidFromOk returns a tuple with the ValidFrom field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BusinessCalendarRequest) GetValidFromOk() (*string, bool) {
	if o == nil || IsNil(o.ValidFrom) {
		return nil, false
	}
	return o.ValidFrom, true
}

// HasValidFrom returns a boolean if a field has been set.
func (o *BusinessCalendarRequest) HasValidFrom() bool {
	if o != nil && !IsNil(o.ValidFrom) {
		return true
	}

	return false
}

// SetValidFrom gets a reference to the given string and assigns it to the ValidFrom field.
func (o *BusinessCalendarRequest) SetValidFrom(v string) {
	o.ValidFrom = &v
}

// GetValidTo returns the ValidTo field value if set, zero value otherwise.
func (o *BusinessCalendarRequest) GetValidTo() string {
	if o == nil || IsNil(o.ValidTo) {
		var ret string
		return ret
	}
	return *o.ValidTo
}

// GetValidToOk returns a tuple with the ValidTo field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BusinessCalendarRequest) GetValidToOk() (*string, bool) {
	if o == nil || IsNil(o.ValidTo) {
		return nil, false
	}
	return o.ValidTo, true
}

// HasValidTo returns a boolean if a field has been set.
func (o *BusinessCalendarRequest) HasValidTo() bool {
	if o != nil && !IsNil(o.ValidTo) {
		return true
	}

	return false
}

// SetValidTo gets a reference to the given string and assigns it to the ValidTo field.
func (o *BusinessCalendarRequest) SetValidTo(v string) {
	o.ValidTo = &v
}

// GetDescription returns the Description field value if set, zero value otherwise.
func (o *BusinessCalendarRequest) GetDescription() string {
	if o == nil || IsNil(o.Description) {
		var ret string
		return ret
	}
	return *o.Description
}

// GetDescriptionOk returns a tuple with the Description field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *BusinessCalendarRequest) GetDescriptionOk() (*string, bool) {
	if o == nil || IsNil(o.Description) {
		return nil, false
	}
	return o.Description, true
}

// HasDescription returns a boolean if a field has been set.
func (o *BusinessCalendarRequest) HasDescription() bool {
	if o != nil && !IsNil(o.Description) {
		return true
	}

	return false
}

// SetDescription gets a reference to the given string and assigns it to the Description field.
func (o *BusinessCalendarRequest) SetDescription(v string) {
	o.Description = &v
}

func (o BusinessCalendarRequest) MarshalJSON() ([]byte, error) {
	toSerialize, err := o.ToMap()
	if err != nil {
		return []byte{}, err
	}
	return json.Marshal(toSerialize)
}

func (o BusinessCalendarRequest) ToMap() (map[string]interface{}, error) {
	toSerialize := map[string]interface{}{}
	if !IsNil(o.Id) {
		toSerialize["id"] = o.Id
	}
	toSerialize["title"] = o.Title
	if !IsNil(o.Weekstart) {
		toSerialize["weekstart"] = o.Weekstart
	}
	if !IsNil(o.Weekdays) {
		toSerialize["weekdays"] = o.Weekdays
	}
	if !IsNil(o.Holidays) {
		toSerialize["holidays"] = o.Holidays
	}
	if !IsNil(o.ValidFrom) {
		toSerialize["validFrom"] = o.ValidFrom
	}
	if !IsNil(o.ValidTo) {
		toSerialize["validTo"] = o.ValidTo
	}
	if !IsNil(o.Description) {
		toSerialize["description"] = o.Description
	}
	return toSerialize, nil
}

type NullableBusinessCalendarRequest struct {
	value *BusinessCalendarRequest
	isSet bool
}

func (v NullableBusinessCalendarRequest) Get() *BusinessCalendarRequest {
	return v.value
}

func (v *NullableBusinessCalendarRequest) Set(val *BusinessCalendarRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableBusinessCalendarRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableBusinessCalendarRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableBusinessCalendarRequest(val *BusinessCalendarRequest) *NullableBusinessCalendarRequest {
	return &NullableBusinessCalendarRequest{value: val, isSet: true}
}

func (v NullableBusinessCalendarRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableBusinessCalendarRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}
