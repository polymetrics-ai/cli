import type { Metadata } from 'next';
import { BookmarksView } from '@/components/blog/bookmarks-view';

export const metadata: Metadata = {
  title: 'Bookmarks',
  description: 'Your saved passages from the Polymetrics blog.',
  robots: { index: false, follow: false },
};

export default function BookmarksPage() {
  return <BookmarksView />;
}
