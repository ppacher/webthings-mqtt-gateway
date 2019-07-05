package spec

import (
	"encoding/json"
	"net/http"

	"github.com/own-home/central/pkg/errors"
)

var (
	// ErrMissingThingID indicates that the thing has now ID specified
	ErrMissingThingID = errors.NewWithStatus(http.StatusBadRequest, "thing ID is missing")

	// ErrMissingItemID indicates that the item has now ID specified
	ErrMissingItemID = errors.NewWithStatus(http.StatusBadRequest, "item ID is missing")

	// ErrInvalidKind indicates that the item kind is unknown or invalid
	ErrInvalidKind = errors.NewWithStatus(http.StatusBadRequest, "invalid or unknown item kind")

	// ErrReadonlyItemWithSet indicates that the item has it's Readonly field set but SetTopic or SetPayload
	// has been configured as well
	ErrReadonlyItemWithSet = errors.NewWithStatus(http.StatusBadRequest, "item marked as readonly but set topic/payload configured")

	// ErrInvalidPayloadHandler indicates that the payload handler configured is unknown
	ErrInvalidPayloadHandler = errors.NewWithStatus(http.StatusBadRequest, "invalid payload handler")
)

// ValidationError wraps a set of error messages that were found when
// validating something
type ValidationError struct {
	Errors []error
}

// Add a new validation error
func (v *ValidationError) Add(err error) {
	v.Errors = append(v.Errors, err)
}

// Len returns the number of validation errors
func (v *ValidationError) Len() int {
	return len(v.Errors)
}

// Error implements the error interface for ValidationError
func (v *ValidationError) Error() string {
	if v == nil || len(v.Errors) == 0 {
		return "<noerror>"
	}

	var str string
	for _, err := range v.Errors {
		if str != "" {
			str = str + "; "
		}
		str = str + err.Error()
	}
	return str
}

func (v *ValidationError) StatusCode() int {
	return 400
}

func (v *ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"code":  v.StatusCode(),
		"error": v.Error(),
	})
}

// NewValidationError returns a new validation error wrapping all errors
// provided. If no errors are given NewValidationError returns nil
func NewValidationError(err ...error) *ValidationError {
	if len(err) == 0 {
		return nil
	}

	return &ValidationError{err}
}
