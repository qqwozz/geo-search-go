CREATE INDEX IF NOT EXISTS idx_pois_outlets ON pois (has_outlets) WHERE has_outlets = TRUE;
CREATE INDEX IF NOT EXISTS idx_pois_parking ON pois (has_parking) WHERE has_parking = TRUE;
CREATE INDEX IF NOT EXISTS idx_pois_live_music ON pois (has_live_music) WHERE has_live_music = TRUE;
CREATE INDEX IF NOT EXISTS idx_pois_romantic ON pois (is_romantic) WHERE is_romantic = TRUE;
CREATE INDEX IF NOT EXISTS idx_pois_family ON pois (is_family_friendly) WHERE is_family_friendly = TRUE;
CREATE INDEX IF NOT EXISTS idx_pois_dog ON pois (is_dog_friendly) WHERE is_dog_friendly = TRUE;
