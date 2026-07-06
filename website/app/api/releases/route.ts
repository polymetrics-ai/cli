import { NextResponse } from 'next/server';
import { getGithubReleases } from '@/lib/github-releases';

export const revalidate = 300;

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url);
  const parsedLimit = Number(searchParams.get('limit') ?? 5);
  const limit = Math.min(
    Math.max(Number.isNaN(parsedLimit) ? 5 : parsedLimit, 1),
    20,
  );
  const releases = await getGithubReleases(limit);

  return NextResponse.json({
    repository: 'polymetrics-ai/cli',
    releases,
  });
}
