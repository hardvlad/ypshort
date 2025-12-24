DROP INDEX IF EXISTS idx_saved_links_url;
CREATE UNIQUE INDEX idx_saved_links_url ON saved_links(url); 