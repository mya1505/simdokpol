CREATE TABLE "licenses" (
    "key" VARCHAR(255) PRIMARY KEY,
    "status" VARCHAR(50) NOT NULL,
    "activated_at" TIMESTAMPTZ,
    "activated_by_id" INTEGER,
    "notes" TEXT
);

INSERT INTO "configurations" ("key", "value") 
VALUES ('license_status', 'UNLICENSED') 
ON CONFLICT ("key") DO NOTHING;