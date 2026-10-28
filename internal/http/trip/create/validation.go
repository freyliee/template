package create

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	api "github.com/freyliee/template/internal/generated"
)

func decodeCreateTrip(
	body io.Reader,
) (api.CreateTripJSONRequestBody, error) {
	var data api.CreateTripJSONRequestBody
	var raw json.RawMessage

	decoder := json.NewDecoder(body)

	if err := decoder.Decode(&raw); err != nil {
		return data, fmt.Errorf("request body must contain valid JSON")
	}

	var extra json.RawMessage

	if err := decoder.Decode(&extra); err != io.EOF {
		return data, fmt.Errorf(
			"request body must contain exactly one JSON value",
		)
	}

	if err := validateRequiredTripFields(raw); err != nil {
		return data, err
	}

	typedDecoder := json.NewDecoder(bytes.NewReader(raw))
	typedDecoder.DisallowUnknownFields()

	if err := typedDecoder.Decode(&data); err != nil {
		return data, fmt.Errorf(
			"request body contains invalid values or unknown fields",
		)
	}

	return data, nil
}

func validateRequiredTripFields(body json.RawMessage) error {
	fields, err := decodeObject(body)
	if err != nil {
		return fmt.Errorf("request body must be an object")
	}

	for _, name := range []string{
		"user_id",
		"driver_id",
		"start_point",
		"end_point",
		"price",
	} {
		if err := requireField(fields, name); err != nil {
			return err
		}
	}

	for _, name := range []string{"start_point", "end_point"} {
		point, err := decodeObject(fields[name])
		if err != nil {
			return fmt.Errorf("%s must be an object", name)
		}

		for _, coordinate := range []string{"latitude", "longitude"} {
			if err := requireField(point, coordinate); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	}

	return nil
}

func decodeObject(
	body json.RawMessage,
) (map[string]json.RawMessage, error) {
	var fields map[string]json.RawMessage

	if err := json.Unmarshal(body, &fields); err != nil {
		return nil, err
	}

	if fields == nil {
		return nil, fmt.Errorf("expected an object")
	}

	return fields, nil
}

func requireField(
	fields map[string]json.RawMessage,
	name string,
) error {
	value, exists := fields[name]

	if !exists ||
		bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
		return fmt.Errorf("%s is required and must not be null", name)
	}

	return nil
}
