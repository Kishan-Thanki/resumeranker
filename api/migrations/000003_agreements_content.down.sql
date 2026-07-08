ALTER TABLE agreements DROP COLUMN content;
ALTER TABLE agreements ADD COLUMN document_url TEXT NOT NULL DEFAULT '';
