package handler

import (
	"context"
	"net/http"
	"time"

	api "github.com/freyliee/template/internal/generated"
	"github.com/freyliee/template/internal/model"
	"github.com/google/uuid"
)

func (h *Handler) CreateTrip(
	w http.ResponseWriter,
	r *http.Request,
	_ api.CreateTripParams,
) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	r = r.WithContext(ctx)

	data, err := decodeCreateTrip(r.Body)
	if err != nil {
		writeProblem(
			w, r,
			http.StatusBadRequest,
			"invalid_request",
			"Invalid request",
			err.Error(),
		)
		return
	}

	trip := api.Trip{
		Id:         uuid.New(),
		UserId:     data.UserId,
		DriverId:   data.DriverId,
		StartPoint: data.StartPoint,
		EndPoint:   data.EndPoint,
		Price:      data.Price,
		Status:     api.Active,
		StartedAt:  time.Now().UTC(),
	}

	err = h.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := h.trips.CreateTrip(txCtx, trip); err != nil {
			return err
		}

		return h.trips.AddStatusHistory(
			txCtx,
			trip.Id,
			nil,
			api.Active,
			"trip created",
		)
	})
	if err != nil {
		writeTripError(w, r, err)
		return
	}

	w.Header().Set(
		"Location",
		"/api/v1/trips/"+trip.Id.String(),
	)

	writeJSON(
		w, r,
		http.StatusCreated,
		"application/json",
		trip,
	)
}

func (h *Handler) GetTrip(
	w http.ResponseWriter,
	r *http.Request,
	tripId api.TripId,
) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	r = r.WithContext(ctx)

	trip, err := h.trips.GetTrip(ctx, tripId)
	if err != nil {
		writeTripError(w, r, err)
		return
	}

	writeJSON(
		w, r,
		http.StatusOK,
		"application/json",
		trip,
	)
}

func (h *Handler) FinishTrip(
	w http.ResponseWriter,
	r *http.Request,
	tripId api.TripId,
) {
	ctx, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	r = r.WithContext(ctx)

	err := h.txManager.Do(ctx, func(txCtx context.Context) error {
		trip, err := h.trips.GetTripForUpdate(txCtx, tripId)
		if err != nil {
			return err
		}

		if trip.Status == api.Completed {
			return model.ErrTripCompleted
		}

		finishedAt := time.Now().UTC()

		if err := h.trips.FinishTrip(
			txCtx,
			tripId,
			finishedAt,
		); err != nil {
			return err
		}

		return h.trips.AddStatusHistory(
			txCtx,
			tripId,
			&trip.Status,
			api.Completed,
			"trip finished",
		)
	})
	if err != nil {
		writeTripError(w, r, err)
		return
	}

	h.GetTrip(w, r, tripId)
}
