ALTER TABLE agreements DROP COLUMN document_url;
ALTER TABLE agreements ADD COLUMN content TEXT NOT NULL DEFAULT '';
