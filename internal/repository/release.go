package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/wuge-xu/shipguard/internal/release"
)

var ErrReleaseNotFound = errors.New("release not found")

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
