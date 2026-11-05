package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	api "github.com/freyliee/template/internal/generated"
	"github.com/freyliee/template/internal/model"
)

func writeProblem(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	code string,
	title string,
	detail string,
) {
	instance := r.URL.Path

	response := api.Problem{
		Type: "https://tripgo.example/problems/" +
			strings.ReplaceAll(code, "_", "-"),
		Title:    title,
		Status:   int32(status),
		Code:     code,
		Instance: &instance,
	}

	if detail != "" {
		response.Detail = &detail
	}

	writeJSON(
		w,
		r,
		status,
		"application/problem+json",
		response,
	)
}

func writeJSON(
	w http.ResponseWriter,
	r *http.Request,
	status int,
	contentType string,
	value any,
) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.ErrorContext(
			r.Context(),
			"write HTTP response",
			"error", err,
		)
	}
}

func writeTripError(
	w http.ResponseWriter,
	r *http.Request,
	err error,
) {
	switch {
	case errors.Is(err, model.ErrTripNotFound):
		writeProblem(
			w, r,
			http.StatusNotFound,
			"trip_not_found",
			"Trip not found",
			"Trip was not found",
		)

	case errors.Is(err, model.ErrTripCompleted):
		writeProblem(
			w, r,
			http.StatusConflict,
			"trip_completed",
			"Trip completed",
			"Operation is not allowed for a completed trip",
		)

	case errors.Is(err, model.ErrDriverBusy):
		writeProblem(
			w, r,
			http.StatusConflict,
			"driver_busy",
			"Driver busy",
			"Driver already has an active trip",
		)

	default:
		slog.ErrorContext(
			r.Context(),
			"trip operation failed",
			"error", err,
		)

		writeProblem(
			w, r,
			http.StatusInternalServerError,
			"internal_error",
			"Internal Server Error",
			"Internal server error",
		)
	}
}

func HandleAPIError(
	w http.ResponseWriter,
	r *http.Request,
	_ error,
) {
	writeProblem(
		w, r,
		http.StatusBadRequest,
		"invalid_request",
		"Invalid request",
		"Request parameter has an invalid format",
	)
}
