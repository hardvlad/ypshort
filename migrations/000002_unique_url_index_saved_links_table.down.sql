DROP INDEX IF EXISTS idx_saved_links_url;
CREATE INDEX idx_saved_links_url ON saved_links(url); 