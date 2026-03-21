BEGIN TRANSACTION;

CREATE TABLE IF NOT EXISTS RainGaugeDevice(
    id                INTEGER PRIMARY KEY,
    id_in_arcgis      TEXT NOT NULL,
    name              TEXT NOT NULL,
    time_of_reading   DATETIME NOT NULL,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at        DATETIME
);

CREATE TABLE IF NOT EXISTS RainGaugeReadings(
    id              INTEGER PRIMARY KEY,
    reading_value   FLOAT NOT NULL,
    device_id       INTEGER NOT NULL,
    created_at      DATETIME DEFAULT CURRENT_TIMESTAMPT NOT NULL,
    deleted_at      DATETIME
);

CREATE TABLE IF NOT EXISTS Camera(
    id                INTEGER PRIMARY KEY,
    name              TEXT NOT NULL,
    id_in_nittrans    TEXT NOT NULL,
    code_in_nittrans  TEXT NOT NULL,
    latitude          TEXT NOT NULL,
    longitude         TEXT NOT NULL,
    created_at        DATETIME DEFAULT CURRENT_TIMESTAMPT NOT NULL,
    deleted_at        DATETIME
);

CREATE TABLE IF NOT EXISTS CameraSnapshot(
    id           INTEGER PRIMARY KEY,
    image_url    TEXT NOT NULL,
    from_camera  INTEGER NOT NULL,
    created_at   DATETIME DEFAULT CURRENT_TIMESTAMPT NOT NULL,
    deleted_at   DATETIME
);

COMMIT;
