CREATE TABLE releases (
    id TEXT PRIMARY KEY,

    service TEXT NOT NULL,
    environment TEXT NOT NULL,
    source_sha TEXT NOT NULL,
    image_digest TEXT NOT NULL,
    gitops_sha TEXT NOT NULL DEFAULT '',

    status TEXT NOT NULL DEFAULT 'pending_approval',
    version BIGINT NOT NULL DEFAULT 1,

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    CONSTRAINT releases_id_not_blank
        CHECK (BTRIM(id) <> ''),

    CONSTRAINT releases_service_not_blank
        CHECK (BTRIM(service) <> ''),

    CONSTRAINT releases_environment_not_blank
        CHECK (BTRIM(environment) <> ''),

    CONSTRAINT releases_source_sha_not_blank
        CHECK (BTRIM(source_sha) <> ''),

    CONSTRAINT releases_image_digest_not_blank
        CHECK (BTRIM(image_digest) <> ''),

    CONSTRAINT releases_status_valid
        CHECK (
            status IN (
                'pending_approval',
                'approved',
                'deploying',
                'verifying',
                'succeeded',
                'failed',
                'rolling_back',
                'rolled_back',
                'canceled'
            )
        ),

    CONSTRAINT releases_version_positive
        CHECK (version >= 1),

    CONSTRAINT releases_updated_at_valid
        CHECK (updated_at >= created_at)
);

CREATE INDEX releases_status_created_at_idx
    ON releases (status, created_at);

CREATE INDEX releases_service_environment_created_at_idx
    ON releases (service, environment, created_at DESC);