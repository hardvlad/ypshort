create table saved_links (
	code varchar(20), 
	url varchar(500)
);

CREATE INDEX idx_saved_links_code ON saved_links(code);
CREATE INDEX idx_saved_links_url ON saved_links(url); 