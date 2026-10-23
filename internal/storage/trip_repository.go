package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	api "github.com/freyliee/template/internal/generated"
	"github.com/freyliee/template/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TripRepository struct {
	pool *pgxpool.Pool
}

func NewTripRepository(pool *pgxpool.Pool) *TripRepository {
	return &TripRepository{pool: pool}
}

func (t *TripRepository) GetTrip(ctx context.Context, tripId api.TripId) (api.Trip, error) {
	return t.getTrip(ctx, tripId, false)
}

func (t *TripRepository) GetTripForUpdate(ctx context.Context, tripId api.TripId) (api.Trip, error) {
	return t.getTrip(ctx, tripId, true)
}

func (t *TripRepository) CreateTrip(ctx context.Context, trip api.Trip) error {
	executor := executorFromContext(ctx, t.pool)

	query, args, err := sq.
		Insert("trips").
		Columns(
			"id",
			"user_id",
			"driver_id",
			"start_latitude",
			"start_longitude",
			"end_latitude",
			"end_longitude",
			"price",
			"status",
			"started_at",
		).
		Values(
			trip.Id,
			trip.UserId,
			trip.DriverId,
			trip.StartPoint.Latitude,
			trip.StartPoint.Longitude,
			trip.EndPoint.Latitude,
			trip.EndPoint.Longitude,
			trip.Price,
			trip.Status,
			trip.StartedAt,
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("build create trip query: %w", err)
	}

	_, err = executor.Exec(ctx, query, args...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) &&
			pgErr.Code == "23505" &&
			pgErr.ConstraintName == "trips_one_active_per_driver_idx" {
			return model.ErrDriverBusy
		}

		return fmt.Errorf("create trip: %w", err)
	}

	return nil

	return nil
}

func (t *TripRepository) AddStatusHistory(ctx context.Context, tripId api.TripId, fromStatus *api.TripStatus, toStatus api.TripStatus, reason string) error {
	executor := executorFromContext(ctx, t.pool)

	query, args, err := sq.
		Insert("trip_status_history").
		Columns(
			"trip_id",
			"from_status",
			"to_status",
			"reason",
		).
		Values(
			tripId,
			fromStatus,
			toStatus,
			reason,
		).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return fmt.Errorf("build add status history query: %w", err)
	}

	_, err = executor.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("add status history: %w", err)
	}

	return nil
}

func (t *TripRepository) FinishTrip(ctx context.Context, tripId api.TripId, finishedAt time.Time) error {
	executor := executorFromContext(ctx, t.pool)

	query, args, err := sq.
		Update("trips").
		Set("status", api.Completed).
		Set("finished_at", finishedAt).
		Set("updated_at", finishedAt).
		Where(sq.Eq{"id": tripId}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		return fmt.Errorf("build finish trip query: %w", err)
	}

	_, err = executor.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("finish trip: %w", err)
	}

	return nil
}

func (t *TripRepository) getTrip(ctx context.Context, tripId api.TripId, forUpdate bool) (api.Trip, error) {
	executor := executorFromContext(ctx, t.pool)

	builder := sq.Select(
		"id",
		"user_id",
		"driver_id",
		"start_latitude",
		"start_longitude",
		"end_latitude",
		"end_longitude",
		"price",
		"status",
		"started_at",
		"finished_at",
	).
		From("trips").
		Where(sq.Eq{"id": tripId}).
		PlaceholderFormat(sq.Dollar)

	if forUpdate {
		builder.Suffix("FOR UPDATE")
	}

	query, args, err := builder.ToSql()

	if err != nil {
		return api.Trip{}, fmt.Errorf("build get trip query: %w", err)
	}

	var trip api.Trip
	err = executor.QueryRow(ctx, query, args...).Scan(
		&trip.Id,
		&trip.UserId,
		&trip.DriverId,
		&trip.StartPoint.Latitude,
		&trip.StartPoint.Longitude,
		&trip.EndPoint.Latitude,
		&trip.EndPoint.Longitude,
		&trip.Price,
		&trip.Status,
		&trip.StartedAt,
		&trip.FinishedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.Trip{}, model.ErrTripNotFound
		}

		return api.Trip{}, fmt.Errorf("get trip: %w", err)
	}

	return trip, nil
}
