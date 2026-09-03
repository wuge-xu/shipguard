package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wuge-xu/shipguard/internal/release"
)

var (
	ErrReleaseNotFound        = errors.New("release not found")
	ErrReleaseVersionConflict = errors.New("release version conflict")
)

type ReleaseRepository struct {
	pool *pgxpool.Pool
}

func NewReleaseRepository(
	pool *pgxpool.Pool,
) *ReleaseRepository {
	return &ReleaseRepository{
		pool: pool,
	}
}

func (r *ReleaseRepository) Create(
	ctx context.Context,
	item release.Release,
) error {
	if err := item.Validate(); err != nil {
		return fmt.Errorf(
			"validate release: %w",
			err,
		)
	}

	_, err := r.pool.Exec(
		ctx,
		`
		INSERT INTO releases (
			id,
			service,
			environment,
			source_sha,
			image_digest,
			gitops_sha,
			status,
			version,
			created_at,
			updated_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8,
			$9,
			$10
		)
		`,
		item.ID,
		item.Service,
		item.Environment,
		item.SourceSHA,
		item.ImageDigest,
		item.GitOpsSHA,
		string(item.Status),
		item.Version,
		item.CreatedAt,
		item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"insert release %q: %w",
			item.ID,
			err,
		)
	}

	return nil
}

func (r *ReleaseRepository) GetByID(
	ctx context.Context,
	id string,
) (release.Release, error) {
	var item release.Release
	var status string

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT
			id,
			service,
			environment,
			source_sha,
			image_digest,
			gitops_sha,
			status,
			version,
			created_at,
			updated_at
		FROM releases
		WHERE id = $1
		`,
		id,
	).Scan(
		&item.ID,
		&item.Service,
		&item.Environment,
		&item.SourceSHA,
		&item.ImageDigest,
		&item.GitOpsSHA,
		&status,
		&item.Version,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return release.Release{}, fmt.Errorf(
			"%w: %s",
			ErrReleaseNotFound,
			id,
		)
	}

	if err != nil {
		return release.Release{}, fmt.Errorf(
			"query release %q: %w",
			id,
			err,
		)
	}

	item.Status = release.Status(status)

	if err := item.Validate(); err != nil {
		return release.Release{}, fmt.Errorf(
			"validate stored release %q: %w",
			id,
			err,
		)
	}

	return item, nil
}

func (r *ReleaseRepository) UpdateTransition(
	ctx context.Context,
	current release.Release,
	next release.Release,
) error {
	if err := current.Validate(); err != nil {
		return fmt.Errorf(
			"validate current release: %w",
			err,
		)
	}

	if err := next.Validate(); err != nil {
		return fmt.Errorf(
			"validate next release: %w",
			err,
		)
	}

	if current.ID != next.ID {
		return fmt.Errorf(
			"%w: release ID changed from %q to %q",
			release.ErrInvalidRelease,
			current.ID,
			next.ID,
		)
	}

	if next.Version != current.Version+1 {
		return fmt.Errorf(
			"%w: version must advance from %d to %d",
			release.ErrInvalidRelease,
			current.Version,
			current.Version+1,
		)
	}

	if err := release.ValidateTransition(
		current.Status,
		next.Status,
	); err != nil {
		return fmt.Errorf(
			"validate persisted transition: %w",
			err,
		)
	}

	result, err := r.pool.Exec(
		ctx,
		`
		UPDATE releases
		SET
			status = $1,
			version = $2,
			updated_at = $3
		WHERE id = $4
		  AND version = $5
		`,
		string(next.Status),
		next.Version,
		next.UpdatedAt,
		current.ID,
		current.Version,
	)
	if err != nil {
		return fmt.Errorf(
			"update release %q: %w",
			current.ID,
			err,
		)
	}

	if result.RowsAffected() == 1 {
		return nil
	}

	exists, err := r.exists(
		ctx,
		current.ID,
	)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf(
			"%w: %s",
			ErrReleaseNotFound,
			current.ID,
		)
	}

	return fmt.Errorf(
		"%w: release %s expected version %d",
		ErrReleaseVersionConflict,
		current.ID,
		current.Version,
	)
}

func (r *ReleaseRepository) exists(
	ctx context.Context,
	id string,
) (bool, error) {
	var exists bool

	err := r.pool.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM releases
			WHERE id = $1
		)
		`,
		id,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf(
			"check release %q existence: %w",
			id,
			err,
		)
	}

	return exists, nil
}
