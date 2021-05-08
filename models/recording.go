// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// Recording recording
//
// swagger:model recording
type Recording struct {

	// ID of Recording
	ID string `json:"id,omitempty"`

	// Id of Meeting that is Recorded
	MeetingID string `json:"meeting_id,omitempty"`
}

// Validate validates this recording
func (m *Recording) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *Recording) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Recording) UnmarshalBinary(b []byte) error {
	var res Recording
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}