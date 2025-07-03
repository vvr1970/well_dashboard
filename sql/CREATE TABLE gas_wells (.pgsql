CREATE TABLE gas_wells (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(200) NOT NULL,
    production FLOAT,
    status VARCHAR(50) CHECK (status IN ('active', 'inactive', 'maintenance'))
);

ALTER SEQUENCE gas_wells_id_seq RESTART WITH 1000000;