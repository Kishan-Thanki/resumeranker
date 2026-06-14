import { db } from './src/lib/server/db/index';
import { pages, contentBlocks } from './src/lib/server/db/schema';
import * as dotenv from 'dotenv';
dotenv.config();

async function seed() {
	try {
		console.log('Seeding database...');
		
		// Insert landing page
		await db.insert(pages).values({
			slug: 'landing',
			title: 'Resume Ranker',
			metaDescription: 'AI powered resume analysis.'
		}).onConflictDoNothing();

		await db.insert(contentBlocks).values([
			{ id: 'landing-hero_title', pageSlug: 'landing', blockKey: 'hero_title', contentValue: 'Get Your Resume Ranked by AI in Seconds' },
			{ id: 'landing-hero_subtitle', pageSlug: 'landing', blockKey: 'hero_subtitle', contentValue: 'Upload your resume and get instant feedback, tailored specifically for top tech companies.' }
		]).onConflictDoNothing();

		// Insert about page
		await db.insert(pages).values({
			slug: 'about',
			title: 'About Us',
			metaDescription: 'Learn more about Resume Ranker'
		}).onConflictDoNothing();
		
		await db.insert(contentBlocks).values([
			{ id: 'about-content', pageSlug: 'about', blockKey: 'content', contentValue: '<p>We are a team of AI engineers passionate about helping people get hired.</p>' }
		]).onConflictDoNothing();

		console.log('Seeding complete!');
		process.exit(0);
	} catch (e) {
		console.error(e);
		process.exit(1);
	}
}

seed();
