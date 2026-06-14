import { db } from '$lib/server/db';
import { pages, contentBlocks } from '$lib/server/db/schema';
import { eq } from 'drizzle-orm';
import type { PageServerLoad } from './$types';
import { error } from '@sveltejs/kit';

export const load: PageServerLoad = async () => {
	const pageResult = await db.select().from(pages).where(eq(pages.slug, 'landing'));
	if (pageResult.length === 0) {
		error(404, 'Landing page not found in CMS');
	}

	const blocksResult = await db.select().from(contentBlocks).where(eq(contentBlocks.pageSlug, 'landing'));
	
	// Convert array of blocks into a fast key-value lookup map
	const content = blocksResult.reduce((acc, block) => {
		acc[block.blockKey] = block.contentValue;
		return acc;
	}, {} as Record<string, string>);

	return {
		page: pageResult[0],
		content
	};
};
