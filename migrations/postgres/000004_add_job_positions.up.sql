CREATE TABLE "job_positions" (
    "id" SERIAL PRIMARY KEY,
    "nama" VARCHAR(150) NOT NULL UNIQUE,
    "is_active" BOOLEAN NOT NULL DEFAULT true,
    "created_at" TIMESTAMPTZ,
    "updated_at" TIMESTAMPTZ,
    "deleted_at" TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS "idx_job_positions_deleted_at" ON "job_positions"("deleted_at");
