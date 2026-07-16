package task

import (
	"encoding/json"
	"strings"
)

type ValidationError struct {
	Fields map[string]string `json:"fields"`
}

func (e ValidationError) Error() string { return "validation failed" }

func (e ValidationError) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{
		"error":  "validation failed",
		"fields": e.Fields,
	})
}

func (r CreateRequest) Validate() error {
	fields := map[string]string{}
	title := strings.TrimSpace(r.Title)
	if title == "" {
		fields["title"] = "required"
	} else if len(title) > 200 {
		fields["title"] = "max length is 200"
	}
	if len(fields) > 0 {
		return ValidationError{Fields: fields}
	}
	return nil
}

func (r CreateRequest) NormalizedTitle() string {
	return strings.TrimSpace(r.Title)
}
