CREATE TABLE "users" (
  "id" bigserial PRIMARY KEY,
  "login" varchar,
  "password" varchar,
  "address" varchar,
  "created_at" timestamp NOT NULL DEFAULT 'now()',
  "session" varchar 
);

CREATE TABLE "files" (
  "id" bigserial PRIMARY KEY,
  "owner" bigserial,
  "name" varchar NOT NULL DEFAULT 'New File',
  "location" varchar,
  "location_wav" varchar,
  "location_txt" varchar,
  "created_at" timestamp NOT NULL DEFAULT 'now()',
  "status" varchar NOT NULL DEFAULT 'ACCEPTED',
  "guid" varchar 
);

ALTER TABLE "files" ADD FOREIGN KEY ("owner") REFERENCES "users" ("id");
