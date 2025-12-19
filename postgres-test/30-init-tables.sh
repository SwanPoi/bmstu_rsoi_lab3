#!/usr/bin/env bash
set -e

POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}
PGPASSWORD=$POSTGRES_PASSWORD
export PGPASSWORD

echo "Create cars table"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "cars" <<-EOSQL
    CREATE TABLE IF NOT EXISTS cars
    (
        id                  SERIAL PRIMARY KEY,
        car_uid             uuid UNIQUE NOT NULL,
        brand               VARCHAR(80) NOT NULL,
        model               VARCHAR(80) NOT NULL,
        registration_number VARCHAR(20) NOT NULL,
        power               INT,
        price               INT         NOT NULL,
        type                VARCHAR(20) CHECK (type IN ('SEDAN', 'SUV', 'MINIVAN', 'ROADSTER')),
        availability        BOOLEAN     NOT NULL
    );

    INSERT INTO cars (car_uid, brand, model, registration_number, power, price, type, availability) VALUES
    ('109b42f3-198d-4c89-9276-a7520a7120ab', 'Mercedes Benz', 'GLA 250', 'ЛО777Х799', 249, 3500, 'SEDAN', true)
    ON CONFLICT DO NOTHING;
EOSQL

echo "Create payments table"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "payments" <<-EOSQL
    CREATE TABLE IF NOT EXISTS payment
    (
        id          SERIAL PRIMARY KEY,
        payment_uid uuid        NOT NULL,
        status      VARCHAR(20) NOT NULL CHECK (status IN ('PAID', 'CANCELED')),
        price       INT         NOT NULL
    );
EOSQL

echo "Create rentals table"

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "rentals" <<-EOSQL
    CREATE TABLE IF NOT EXISTS rental
    (
        id          SERIAL PRIMARY KEY,
        rental_uid  uuid UNIQUE              NOT NULL,
        username    VARCHAR(80)              NOT NULL,
        payment_uid uuid                     NOT NULL,
        car_uid     uuid                     NOT NULL,
        date_from   TIMESTAMP WITH TIME ZONE NOT NULL,
        date_to     TIMESTAMP WITH TIME ZONE NOT NULL,
        status      VARCHAR(20)              NOT NULL CHECK (status IN ('IN_PROGRESS', 'FINISHED', 'CANCELED'))
    );
EOSQL
