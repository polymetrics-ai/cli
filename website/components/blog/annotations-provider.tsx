'use client';

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from 'react';
import type { ReactNode } from 'react';
import { getBlogPost } from '@/lib/blog';
import { resolveAnchor } from '@/lib/annotations/anchor';
import type { Anchor, Resolution } from '@/lib/annotations/anchor';
import { useSession } from '@/lib/auth-client';
import { SignInDialog } from '@/components/auth/sign-in-dialog';

export type CommentDto = {
  id: string;
  body: string;
  anchor: Anchor;
  createdAt: string;
  author: { name: string; image: string | null };
  mine: boolean;
  /** True while an optimistic insert is awaiting the server. */
  pending?: boolean;
};

export type BookmarkDto = {
  id: string;
  postSlug: string;
  anchor: Anchor;
  createdAt: string;
};

/** A selection the reader just made, in viewport coordinates. */
export type Draft = {
  anchor: Anchor;
  rect: { top: number; left: number; width: number; height: number };
};

type AnnotationsContextValue = {
  slug: string;
  signedIn: boolean;
  comments: CommentDto[];
  bookmarks: BookmarkDto[];
  resolutions: Map<string, Resolution>;
  loading: boolean;
  draft: Draft | null;
  setDraft: (draft: Draft | null) => void;
  composing: boolean;
  openComposer: () => void;
  closeComposer: () => void;
  submitComment: (body: string) => Promise<boolean>;
  deleteComment: (id: string) => Promise<boolean>;
  toggleBookmark: () => Promise<void>;
  isDraftBookmarked: boolean;
  removeBookmark: (id: string) => Promise<void>;
  activeId: string | null;
  setActiveId: (id: string | null) => void;
  viewerAdmin: boolean;
  sheetOpen: boolean;
  setSheetOpen: (open: boolean) => void;
  hovered: { id: string; rect: DOMRect } | null;
  setHovered: (value: { id: string; rect: DOMRect } | null) => void;
  requestSignIn: () => void;
  announce: (message: string) => void;
  announcement: string;
};

const AnnotationsContext = createContext<AnnotationsContextValue | null>(null);

export function useAnnotations(): AnnotationsContextValue {
  const value = useContext(AnnotationsContext);
  if (!value) throw new Error('useAnnotations must be used inside AnnotationsProvider');
  return value;
}

const DRAFT_STORAGE_KEY = 'pm-annotation-draft';

type StashedDraft = { slug: string; anchor: Anchor; body?: string };

export function stashDraft(stash: StashedDraft): void {
  try {
    sessionStorage.setItem(DRAFT_STORAGE_KEY, JSON.stringify(stash));
  } catch {
    /* storage unavailable — the reader just re-selects */
  }
}

function takeStashedDraft(slug: string): StashedDraft | null {
  try {
    const raw = sessionStorage.getItem(DRAFT_STORAGE_KEY);
    if (!raw) return null;
    sessionStorage.removeItem(DRAFT_STORAGE_KEY);
    const parsed = JSON.parse(raw) as StashedDraft;
    return parsed.slug === slug ? parsed : null;
  } catch {
    return null;
  }
}

function anchorsEqual(a: Anchor, b: Anchor): boolean {
  return (
    a.sectionIndex === b.sectionIndex &&
    a.blockType === b.blockType &&
    a.blockIndex === b.blockIndex &&
    a.startOffset === b.startOffset &&
    a.exact === b.exact
  );
}

export function AnnotationsProvider({ slug, children }: { slug: string; children: ReactNode }) {
  const { data: session } = useSession();
  const signedIn = session !== null && session !== undefined;

  const [comments, setComments] = useState<CommentDto[]>([]);
  const [bookmarks, setBookmarks] = useState<BookmarkDto[]>([]);
  const [loading, setLoading] = useState(true);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [composing, setComposing] = useState(false);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [hovered, setHovered] = useState<{ id: string; rect: DOMRect } | null>(null);
  const [signInOpen, setSignInOpen] = useState(false);
  const [announcement, setAnnouncement] = useState('');
  const [viewerAdmin, setViewerAdmin] = useState(false);
  const [sheetOpen, setSheetOpen] = useState(false);
  const restoredRef = useRef(false);

  useEffect(() => {
    const controller = new AbortController();
    fetch(`/api/comments?slug=${encodeURIComponent(slug)}`, { signal: controller.signal })
      .then((response) => (response.ok ? response.json() : { comments: [] }))
      .then((data: { comments: CommentDto[]; viewer?: { admin: boolean } }) => {
        setComments(data.comments);
        setViewerAdmin(data.viewer?.admin ?? false);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
    return () => controller.abort();
  }, [slug, signedIn]);

  useEffect(() => {
    if (!signedIn) {
      setBookmarks([]);
      return;
    }
    const controller = new AbortController();
    fetch(`/api/bookmarks?slug=${encodeURIComponent(slug)}`, { signal: controller.signal })
      .then((response) => (response.ok ? response.json() : { bookmarks: [] }))
      .then((data: { bookmarks: BookmarkDto[] }) => setBookmarks(data.bookmarks))
      .catch(() => {});
    return () => controller.abort();
  }, [slug, signedIn]);

  // After the OAuth round-trip, restore the stashed selection and reopen
  // the composer so the reader continues exactly where they left off.
  useEffect(() => {
    if (!signedIn || restoredRef.current) return;
    restoredRef.current = true;
    const stashed = takeStashedDraft(slug);
    if (!stashed) return;
    setDraft({ anchor: stashed.anchor, rect: { top: 120, left: 120, width: 0, height: 0 } });
    setComposing(true);
  }, [signedIn, slug]);

  const post = useMemo(() => getBlogPost(slug), [slug]);

  const resolutions = useMemo(() => {
    const map = new Map<string, Resolution>();
    if (!post) return map;
    for (const comment of comments) {
      map.set(comment.id, resolveAnchor(post, comment.anchor));
    }
    for (const bookmark of bookmarks) {
      map.set(bookmark.id, resolveAnchor(post, bookmark.anchor));
    }
    return map;
  }, [post, comments, bookmarks]);

  const announce = useCallback((message: string) => {
    setAnnouncement(message);
    window.setTimeout(() => setAnnouncement(''), 2500);
  }, []);

  const requestSignIn = useCallback(() => setSignInOpen(true), []);

  const openComposer = useCallback(() => {
    if (!draft) return;
    setComposing(true);
  }, [draft]);

  const closeComposer = useCallback(() => {
    setComposing(false);
    setDraft(null);
  }, []);

  const submitComment = useCallback(
    async (body: string): Promise<boolean> => {
      if (!draft) return false;
      const anchor = draft.anchor;
      const temporaryId = `pending-${Date.now()}`;
      const optimistic: CommentDto = {
        id: temporaryId,
        body,
        anchor,
        createdAt: new Date().toISOString(),
        author: { name: session?.user.name ?? 'You', image: session?.user.image ?? null },
        mine: true,
        pending: true,
      };
      setComments((current) => [...current, optimistic]);
      setComposing(false);
      setDraft(null);

      try {
        const response = await fetch('/api/comments', {
          method: 'POST',
          headers: { 'content-type': 'application/json' },
          body: JSON.stringify({ slug, body, anchor }),
        });
        if (!response.ok) throw new Error(String(response.status));
        const { comment } = (await response.json()) as { comment: CommentDto };
        setComments((current) => current.map((c) => (c.id === temporaryId ? comment : c)));
        announce('Note posted');
        return true;
      } catch {
        setComments((current) => current.filter((c) => c.id !== temporaryId));
        // Reopen with the draft intact so nothing typed is lost.
        setDraft({ anchor, rect: { top: 120, left: 120, width: 0, height: 0 } });
        setComposing(true);
        announce('Posting failed — your note is still in the editor');
        return false;
      }
    },
    [draft, slug, session, announce],
  );

  const deleteComment = useCallback(
    async (id: string): Promise<boolean> => {
      const previous = comments;
      setComments((current) => current.filter((c) => c.id !== id));
      const response = await fetch(`/api/comments/${id}`, { method: 'DELETE' });
      if (!response.ok) {
        setComments(previous);
        return false;
      }
      announce('Note deleted');
      return true;
    },
    [comments, announce],
  );

  const isDraftBookmarked = useMemo(() => {
    if (!draft) return false;
    return bookmarks.some((bookmark) => anchorsEqual(bookmark.anchor, draft.anchor));
  }, [draft, bookmarks]);

  const toggleBookmark = useCallback(async () => {
    if (!draft) return;
    const anchor = draft.anchor;
    const existing = bookmarks.find((bookmark) => anchorsEqual(bookmark.anchor, anchor));

    if (existing) {
      setBookmarks((current) => current.filter((b) => b.id !== existing.id));
      const response = await fetch(`/api/bookmarks/${existing.id}`, { method: 'DELETE' });
      if (!response.ok) setBookmarks((current) => [...current, existing]);
      else announce('Bookmark removed');
      return;
    }

    const temporaryId = `pending-${Date.now()}`;
    const optimistic: BookmarkDto = {
      id: temporaryId,
      postSlug: slug,
      anchor,
      createdAt: new Date().toISOString(),
    };
    setBookmarks((current) => [...current, optimistic]);
    try {
      const response = await fetch('/api/bookmarks', {
        method: 'POST',
        headers: { 'content-type': 'application/json' },
        body: JSON.stringify({ slug, anchor }),
      });
      if (!response.ok) throw new Error(String(response.status));
      const { bookmark } = (await response.json()) as { bookmark: BookmarkDto };
      setBookmarks((current) => current.map((b) => (b.id === temporaryId ? bookmark : b)));
      announce('Bookmarked');
    } catch {
      setBookmarks((current) => current.filter((b) => b.id !== temporaryId));
      announce('Bookmarking failed');
    }
  }, [draft, bookmarks, slug, announce]);

  const removeBookmark = useCallback(
    async (id: string) => {
      const existing = bookmarks.find((bookmark) => bookmark.id === id);
      if (!existing) return;
      setBookmarks((current) => current.filter((b) => b.id !== id));
      const response = await fetch(`/api/bookmarks/${id}`, { method: 'DELETE' });
      if (!response.ok) setBookmarks((current) => [...current, existing]);
    },
    [bookmarks],
  );

  const value: AnnotationsContextValue = {
    slug,
    signedIn,
    comments,
    bookmarks,
    resolutions,
    loading,
    draft,
    setDraft,
    composing,
    openComposer,
    closeComposer,
    submitComment,
    deleteComment,
    toggleBookmark,
    isDraftBookmarked,
    removeBookmark,
    activeId,
    setActiveId,
    viewerAdmin,
    sheetOpen,
    setSheetOpen,
    hovered,
    setHovered,
    requestSignIn,
    announce,
    announcement,
  };

  return (
    <AnnotationsContext.Provider value={value}>
      {children}
      <SignInDialog open={signInOpen} onOpenChange={setSignInOpen} />
      <span aria-live="polite" role="status" className="sr-only">
        {announcement}
      </span>
    </AnnotationsContext.Provider>
  );
}
