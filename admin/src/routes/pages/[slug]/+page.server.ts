import { db } from '$lib/server/db';
import { pages, contentBlocks } from '$lib/server/db/schema';
import { eq } from 'drizzle-orm';
import { error, fail } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';

export const load: PageServerLoad = async ({ params }) => {
	const pageSlug = params.slug;

	const pageResult = await db.select().from(pages).where(eq(pages.slug, pageSlug));
	if (pageResult.length === 0) {
		error(404, 'Page not found');
	}

	const blocksResult = await db.select().from(contentBlocks).where(eq(contentBlocks.pageSlug, pageSlug));

	return {
		page: pageResult[0],
		blocks: blocksResult
	};
};

export const actions: Actions = {
	update: async ({ request, params }) => {
		const data = await request.formData();
		const pageSlug = params.slug;
		
		const title = data.get('title') as string;
		const metaDescription = data.get('metaDescription') as string;
		
		try {
			// Update page metadata
			await db.update(pages).set({ title, metaDescription }).where(eq(pages.slug, pageSlug));
			
			// Update content blocks
			for (const [key, value] of data.entries()) {
				if (key.startsWith('block_')) {
					const blockKey = key.replace('block_', '');
					await db.update(contentBlocks)
						.set({ contentValue: value as string })
						.where(eq(contentBlocks.id, `${pageSlug}-${blockKey}`));
				}
			}
			return { success: true };
		} catch (e) {
			console.error(e);
			return fail(500, { message: 'Failed to update content' });
		}
	}
};
