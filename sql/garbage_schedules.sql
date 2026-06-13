CREATE TABLE IF NOT EXISTS garbage_schedules (
    location_id VARCHAR(50) PRIMARY KEY,
    date_zmieszane DATE,
    date_papier DATE,
    date_plastik DATE,
    date_szklo DATE,
    date_bio DATE,
    date_zielone DATE,
    date_bio_restauracyjne DATE,
    date_gabaryty DATE,
    last_update TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);