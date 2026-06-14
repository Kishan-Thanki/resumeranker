import { pgTable, text, timestamp } from 'drizzle-orm/pg-core';

export const pages = pgTable('pages', {
	slug: text('slug').primaryKey(),
	title: text('title').notNull(),
	metaDescription: text('meta_description')
});

export const contentBlocks = pgTable('content_blocks', {
	id: text('id').primaryKey(), // We'll use a simple slug-like key e.g., 'landing-hero-title'
	pageSlug: text('page_slug')
		.references(() => pages.slug, { onDelete: 'cascade' })
		.notNull(),
	blockKey: text('block_key').notNull(), // e.g. 'hero_title'
	contentValue: text('content_value').notNull()
});
