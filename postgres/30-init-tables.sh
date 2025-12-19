#!/usr/bin/env bash
set -e

POSTGRES_USER=${POSTGRES_USER:-postgres}
POSTGRES_PASSWORD=${POSTGRES_PASSWORD:-postgres}
PGPASSWORD=$POSTGRES_PASSWORD
export PGPASSWORD

echo "Создание баз данных и пользователя program..."

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
    ('a1b2c3d4-e5f6-7890-a1b2-c3d4e5f67890', 'Toyota', 'Camry', 'A123BC777', 150, 3500, 'SEDAN', false),
    ('b2c3d4e5-f6a7-8901-b2c3-d4e5f6a78901', 'BMW', 'X5', 'B456DE777', 230, 4500, 'SUV', false),
    ('c3d4e5f6-a7b8-9012-c3d4-e5f6a7b89012', 'Mercedes', 'Vito', 'C789FG777', 140, 3000, 'MINIVAN', false),
    ('d4e5f6a7-b8c9-0123-d4e5-f6a7b8c90123', 'Ferrari', '488 GTB', 'D012HI777', 670, 25000, 'ROADSTER', false),
    ('e5f6a7b8-c9d0-1234-e5f6-a7b8c9d01234', 'Honda', 'Civic', 'E345JK777', 120, 2000, 'SEDAN', true),
    ('f6a7b8c9-d0e1-2345-f6a7-b8c9d0e12345', 'Jeep', 'Wrangler', 'F678LM777', 270, 4000, 'SUV', true),
    ('a7b8c9d0-e1f2-3456-a7b8-c9d0e1f23456', 'Ford', 'Transit', 'G901NP777', 130, 2800, 'MINIVAN', true),
    ('b8c9d0e1-f2a3-4567-b8c9-d0e1f2a34567', 'Porsche', '911', 'H234QR777', 450, 15000, 'ROADSTER', false),
    ('c9d0e1f2-a3b4-5678-c9d0-e1f2a3b45678', 'Nissan', 'Altima', 'I567ST777', 180, 2200, 'SEDAN', true),
    ('d0e1f2a3-b4c5-6789-d0e1-f2a3b4c56789', 'Land Rover', 'Discovery', 'J890UV777', 340, 6000, 'SUV', false)
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

    INSERT INTO payment (payment_uid, status, price) VALUES
    ('a1a2b3c4-d5e6-7890-f1a2-b3c4d5e67890', 'PAID', 14000),
    ('a2b3c4d5-e6f7-8901-f2b3-c4d5e6f78901', 'PAID', 18000),
    ('a3c4d5e6-f7a8-9012-f3c4-d5e6f7a89012', 'PAID', 12000),
    ('a4d5e6f7-a8b9-0123-f4d5-e6f7a8b90123', 'CANCELED', 0),
    ('a5e6f7a8-b9c0-1234-f5e6-f7a8b9c01234', 'PAID', 8000),
    ('a6f7a8b9-c0d1-2345-f6f7-a8b9c0d12345', 'PAID', 16000),
    ('a7a8b9c0-d1e2-3456-f7a8-b9c0d1e23456', 'PAID', 11200),
    ('a8b9c0d1-e2f3-4567-f8b9-c0d1e2f34567', 'PAID', 60000),
    ('a9c0d1e2-f3a4-5678-f9c0-d1e2f3a45678', 'PAID', 8800),
    ('a0d1e2f3-a4b5-6789-f0d1-e2f3a4b56789', 'PAID', 24000)
    ON CONFLICT DO NOTHING;
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

    INSERT INTO rental (rental_uid, username, payment_uid, car_uid, date_from, date_to, status) VALUES
    ('b1a2b3c4-d5e6-7890-d1a2-b3c4d5e67890', 'john_doe', 'a1a2b3c4-d5e6-7890-f1a2-b3c4d5e67890', 'a1b2c3d4-e5f6-7890-a1b2-c3d4e5f67890', '2023-12-01 10:00:00+00', '2023-12-05 10:00:00+00', 'IN_PROGRESS'),
    ('b2b3c4d5-e6f7-8901-d2b3-c4d5e6f78901', 'jane_smith', 'a2b3c4d5-e6f7-8901-f2b3-c4d5e6f78901', 'b2c3d4e5-f6a7-8901-b2c3-d4e5f6a78901', '2023-12-02 14:00:00+00', '2023-12-06 14:00:00+00', 'FINISHED'),
    ('b3c4d5e6-f7a8-9012-d3c4-d5e6f7a89012', 'mike_johnson', 'a3c4d5e6-f7a8-9012-f3c4-d5e6f7a89012', 'c3d4e5f6-a7b8-9012-c3d4-e5f6a7b89012', '2023-12-03 09:00:00+00', '2023-12-07 09:00:00+00', 'IN_PROGRESS'),
    ('b4d5e6f7-a8b9-0123-d4d5-e6f7a8b90123', 'sarah_wilson', 'a4d5e6f7-a8b9-0123-f4d5-e6f7a8b90123', 'd4e5f6a7-b8c9-0123-d4e5-f6a7b8c90123', '2023-12-04 16:00:00+00', '2023-12-08 16:00:00+00', 'CANCELED'),
    ('b5e6f7a8-b9c0-1234-d5e6-f7a8b9c01234', 'david_brown', 'a5e6f7a8-b9c0-1234-f5e6-f7a8b9c01234', 'e5f6a7b8-c9d0-1234-e5f6-a7b8c9d01234', '2023-12-05 11:00:00+00', '2023-12-09 11:00:00+00', 'IN_PROGRESS'),
    ('b6f7a8b9-c0d1-2345-d6f7-a8b9c0d12345', 'emily_davis', 'a6f7a8b9-c0d1-2345-f6f7-a8b9c0d12345', 'f6a7b8c9-d0e1-2345-f6a7-b8c9d0e12345', '2023-12-06 13:00:00+00', '2023-12-10 13:00:00+00', 'FINISHED'),
    ('b7a8b9c0-d1e2-3456-d7a8-b9c0d1e23456', 'chris_miller', 'a7a8b9c0-d1e2-3456-f7a8-b9c0d1e23456', 'a7b8c9d0-e1f2-3456-a7b8-c9d0e1f23456', '2023-12-07 15:00:00+00', '2023-12-11 15:00:00+00', 'IN_PROGRESS'),
    ('b8b9c0d1-e2f3-4567-d8b9-c0d1e2f34567', 'laura_garcia', 'a8b9c0d1-e2f3-4567-f8b9-c0d1e2f34567', 'b8c9d0e1-f2a3-4567-b8c9-d0e1f2a34567', '2023-12-08 10:00:00+00', '2023-12-12 10:00:00+00', 'FINISHED'),
    ('b9c0d1e2-f3a4-5678-d9c0-d1e2f3a45678', 'kevin_taylor', 'a9c0d1e2-f3a4-5678-f9c0-d1e2f3a45678', 'c9d0e1f2-a3b4-5678-c9d0-e1f2a3b45678', '2023-12-09 12:00:00+00', '2023-12-13 12:00:00+00', 'IN_PROGRESS'),
    ('b0d1e2f3-a4b5-6789-d0d1-e2f3a4b56789', 'amanda_lee', 'a0d1e2f3-a4b5-6789-f0d1-e2f3a4b56789', 'd0e1f2a3-b4c5-6789-d0e1-f2a3b4c56789', '2023-12-10 14:00:00+00', '2023-12-14 14:00:00+00', 'FINISHED')
    ON CONFLICT DO NOTHING;
EOSQL
