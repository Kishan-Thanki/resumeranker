import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './schema';
import * as dotenv from 'dotenv';
import { building } from '$app/environment';

dotenv.config();

if (!process.env.DATABASE_URL && !building) {
	throw new Error('DATABASE_URL is not set in the environment variables');
}

const client = postgres(process.env.DATABASE_URL || 'postgres://dummy:dummy@localhost:5432/dummy');
export const db = drizzle(client, { schema });
