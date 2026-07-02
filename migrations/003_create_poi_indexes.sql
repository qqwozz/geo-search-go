CREATE INDEX idx_pois_geom ON pois USING GIST (geom);
CREATE INDEX idx_pois_category ON pois (category);
CREATE INDEX idx_pois_rating ON pois (rating DESC);
CREATE INDEX idx_pois_city ON pois (city);
CREATE INDEX idx_pois_wifi ON pois (has_wifi) WHERE has_wifi = TRUE;
CREATE INDEX idx_pois_terrace ON pois (has_terrace) WHERE has_terrace = TRUE;
CREATE INDEX idx_pois_quiet ON pois (is_quiet) WHERE is_quiet = TRUE;
CREATE INDEX idx_pois_breakfast ON pois (has_breakfast) WHERE has_breakfast = TRUE;
