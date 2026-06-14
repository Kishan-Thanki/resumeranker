import { db } from '$lib/server/db';
import { pages } from '$lib/server/db/schema';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async () => {
	const allPages = await db.select().from(pages);
	return { pages: allPages };
};
