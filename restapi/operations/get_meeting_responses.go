// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"github.com/prota-studios/javis/models"
)

// GetMeetingOKCode is the HTTP code returned for type GetMeetingOK
const GetMeetingOKCode int = 200

/*GetMeetingOK Get a given Meeting

swagger:response getMeetingOK
*/
type GetMeetingOK struct {

	/*
	  In: Body
	*/
	Payload *models.Meeting `json:"body,omitempty"`
}

// NewGetMeetingOK creates GetMeetingOK with default headers values
func NewGetMeetingOK() *GetMeetingOK {

	return &GetMeetingOK{}
}

// WithPayload adds the payload to the get meeting o k response
func (o *GetMeetingOK) WithPayload(payload *models.Meeting) *GetMeetingOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get meeting o k response
func (o *GetMeetingOK) SetPayload(payload *models.Meeting) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetMeetingOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
