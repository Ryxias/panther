// Code generated by go-swagger; DO NOT EDIT.

package operations

/**
 * Panther is a scalable, powerful, cloud-native SIEM written in Golang/React.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	strfmt "github.com/go-openapi/strfmt"

	models "github.com/panther-labs/panther/api/gateway/analysis/models"
)

// BulkUploadReader is a Reader for the BulkUpload structure.
type BulkUploadReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *BulkUploadReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewBulkUploadOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewBulkUploadBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewBulkUploadInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewBulkUploadOK creates a BulkUploadOK with default headers values
func NewBulkUploadOK() *BulkUploadOK {
	return &BulkUploadOK{}
}

/*BulkUploadOK handles this case with default header values.

OK
*/
type BulkUploadOK struct {
	Payload *models.BulkUploadResult
}

func (o *BulkUploadOK) Error() string {
	return fmt.Sprintf("[POST /upload][%d] bulkUploadOK  %+v", 200, o.Payload)
}

func (o *BulkUploadOK) GetPayload() *models.BulkUploadResult {
	return o.Payload
}

func (o *BulkUploadOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.BulkUploadResult)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewBulkUploadBadRequest creates a BulkUploadBadRequest with default headers values
func NewBulkUploadBadRequest() *BulkUploadBadRequest {
	return &BulkUploadBadRequest{}
}

/*BulkUploadBadRequest handles this case with default header values.

Bad request
*/
type BulkUploadBadRequest struct {
	Payload *models.Error
}

func (o *BulkUploadBadRequest) Error() string {
	return fmt.Sprintf("[POST /upload][%d] bulkUploadBadRequest  %+v", 400, o.Payload)
}

func (o *BulkUploadBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *BulkUploadBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewBulkUploadInternalServerError creates a BulkUploadInternalServerError with default headers values
func NewBulkUploadInternalServerError() *BulkUploadInternalServerError {
	return &BulkUploadInternalServerError{}
}

/*BulkUploadInternalServerError handles this case with default header values.

Internal server error
*/
type BulkUploadInternalServerError struct {
}

func (o *BulkUploadInternalServerError) Error() string {
	return fmt.Sprintf("[POST /upload][%d] bulkUploadInternalServerError ", 500)
}

func (o *BulkUploadInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
