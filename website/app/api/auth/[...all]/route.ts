import { toNextJsHandler } from 'better-auth/next-js';
import { auth } from '@/lib/auth';
import { ensureMigrated } from '@/lib/db';

export const dynamic = 'force-dynamic';

const handler = toNextJsHandler(auth);

export async function GET(request: Request) {
  await ensureMigrated();
  return handler.GET(request);
}

export async function POST(request: Request) {
  await ensureMigrated();
  return handler.POST(request);
}
