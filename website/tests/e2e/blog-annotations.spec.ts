import { expect, test } from '@playwright/test';
import type { Page } from '@playwright/test';

/**
 * End-to-end coverage for blog annotations (comments + bookmarks).
 * Requires a database; auth uses the double-gated credentials provider
 * (E2E_TEST_AUTH=1, injected by playwright.config.ts webServer env).
 */
const POST_PATH = '/blog/one-cli-to-rule-them-all';
const hasDatabase = Boolean(process.env.DATABASE_URL || process.env.PLAYWRIGHT_BASE_URL);

test.skip(!hasDatabase, 'annotations e2e needs DATABASE_URL (postgres) for the webServer');

/**
 * The navbar session slot renders only after hydration and the session
 * fetch settle — selections dispatched earlier race React's event replay.
 */
async function awaitHydrated(page: Page): Promise<void> {
  await expect(
    page
      .locator('header button[aria-label="Account menu"], header button:has-text("Sign in")')
      .first(),
  ).toBeVisible();
}

async function selectInBlock(page: Page, blockIndex: number, needle: string): Promise<void> {
  await awaitHydrated(page);
  // Real readers select visible text; bring the block on-screen first so
  // the fixed-position popover isn't born outside the viewport, and let
  // the scroll settle so its event doesn't race the selection capture.
  await page.evaluate((index) => {
    document.querySelectorAll('[data-annotation-block]')[index]?.scrollIntoView({ block: 'center' });
  }, blockIndex);
  await page.waitForTimeout(350);
  await page.evaluate(
    ({ blockIndex, needle }) => {
      const block = document.querySelectorAll('[data-annotation-block]')[blockIndex];
      const find = (node: Node): Text | null => {
        for (const child of node.childNodes) {
          if (child.nodeType === Node.TEXT_NODE && (child.textContent ?? '').includes(needle)) {
            return child as Text;
          }
          const nested = find(child);
          if (nested) return nested;
        }
        return null;
      };
      const textNode = find(block);
      if (!textNode) throw new Error(`text "${needle}" not found in block ${blockIndex}`);
      const start = (textNode.textContent ?? '').indexOf(needle);
      const range = document.createRange();
      range.setStart(textNode, start);
      range.setEnd(textNode, start + needle.length);
      const selection = window.getSelection();
      selection?.removeAllRanges();
      selection?.addRange(range);
      document.dispatchEvent(new MouseEvent('mouseup', { bubbles: true }));
    },
    { blockIndex, needle },
  );
}

async function signUp(page: Page, label: string): Promise<void> {
  const email = `e2e-${label}-${Date.now()}@example.com`;
  // Better Auth rate-limits the sign-up path; parallel workers signing up
  // simultaneously can trip it, so retry with a short backoff.
  for (let attempt = 0; attempt < 5; attempt++) {
    const response = await page.request.post('/api/auth/sign-up/email', {
      data: { name: `E2E ${label}`, email, password: 'e2e-password-123' },
    });
    if (response.ok()) return;
    if (response.status() !== 429) break;
    await page.waitForTimeout(3000);
  }
  throw new Error(`sign-up failed for ${email}`);
}

test.describe('blog annotations', () => {
  test('signed-out comment attempt opens the sign-in dialog', async ({ page }) => {
    await page.goto(POST_PATH);
    await selectInBlock(page, 0, 'extract from a source');

    const toolbar = page.getByRole('toolbar', { name: 'Annotate selection' });
    await expect(toolbar).toBeVisible();
    await toolbar.getByRole('button', { name: 'Comment' }).click();

    await expect(page.getByRole('heading', { name: 'Join the discussion' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Continue with GitHub' })).toBeVisible();
  });

  test('signed-in reader comments, sees the highlight and margin note, and deletes it', async ({ page }) => {
    await signUp(page, 'commenter');
    await page.goto(POST_PATH);
    const noteText = `Sharp point — noted at ${Date.now()}`;

    await selectInBlock(page, 0, 'decide what should happen');
    const toolbar = page.getByRole('toolbar', { name: 'Annotate selection' });
    await expect(toolbar).toBeVisible();
    await toolbar.getByRole('button', { name: 'Comment' }).click();

    const composer = page.getByRole('dialog', { name: 'New note' });
    await expect(composer).toBeVisible();
    await composer.getByRole('textbox').fill(noteText);
    await composer.getByRole('button', { name: 'Post note' }).click();

    const mark = page.locator('mark[data-annotation-mark]', { hasText: 'decide what should happen' });
    await expect(mark).toBeVisible();

    // Margin rail (desktop viewport) shows the note aligned to the mark.
    await expect(page.locator('[data-margin-note]').filter({ hasText: noteText })).toBeVisible();

    // Hover preview.
    await mark.hover();
    await expect(page.getByRole('tooltip').filter({ hasText: noteText })).toBeVisible();

    // Delete through the sheet with inline confirm.
    await page.getByRole('button', { name: 'Open all notes' }).click();
    const entry = page.locator('[data-slot=sheet-content] >> text=' + JSON.stringify(noteText));
    await expect(entry).toBeVisible();
    await page.getByRole('button', { name: 'Delete note' }).first().click();
    await page.getByRole('button', { name: 'Confirm' }).click();
    await expect(entry).toBeHidden();
  });

  test('mobile readers reach notes through the floating button and sheet', async ({ page }) => {
    await signUp(page, 'mobile');
    await page.goto(POST_PATH);
    const noteText = `Mobile margin note ${Date.now()}`;

    await selectInBlock(page, 1, 'command-line workflow');
    await page.getByRole('toolbar', { name: 'Annotate selection' }).getByRole('button', { name: 'Comment' }).click();
    const composer = page.getByRole('dialog', { name: 'New note' });
    await composer.getByRole('textbox').fill(noteText);
    await composer.getByRole('button', { name: 'Post note' }).click();
    const mark = page.locator('mark[data-annotation-mark]', { hasText: 'command-line workflow' });
    await expect(mark).toBeVisible();

    await page.setViewportSize({ width: 390, height: 844 });
    await page.getByRole('button', { name: /Open notes/ }).click();
    const sheet = page.locator('[data-slot=sheet-content]');
    const entry = sheet.locator('div', { hasText: noteText }).last();
    await expect(entry).toBeVisible();
    await entry.getByRole('button', { name: 'Jump to text →' }).click();
    await expect(mark).toBeInViewport();
  });

  test('bookmarks toggle, list privately, and deep-link back to the passage', async ({ page }) => {
    await signUp(page, 'bookmarker');
    await page.goto(POST_PATH);

    await selectInBlock(page, 2, 'portable');
    const toolbar = page.getByRole('toolbar', { name: 'Annotate selection' });
    await toolbar.getByRole('button', { name: 'Bookmark' }).click();
    await expect(toolbar.getByRole('button', { name: 'Saved' })).toBeVisible();
    await expect(page.locator('[data-annotation-bookmark]')).toBeVisible();

    await page.goto('/bookmarks');
    const row = page.locator('blockquote', { hasText: 'portable' });
    await expect(row).toBeVisible();

    await page.locator('a[href*="?b="]').first().click();
    await expect(page).toHaveURL(/\?b=/);
    await expect(page.locator('[data-annotation-bookmark]')).toBeInViewport();

    // Remove from the library.
    await page.goto('/bookmarks');
    await page.getByRole('button', { name: 'Remove bookmark' }).first().click();
    await expect(page.getByText('Nothing saved yet')).toBeVisible();
  });

  test('replies thread YouTube-style behind an expander', async ({ page }) => {
    await signUp(page, 'replier');
    await page.goto(POST_PATH);
    const noteText = `Threaded root ${Date.now()}`;
    const replyText = `First reply ${Date.now()}`;

    // Root note.
    await selectInBlock(page, 0, 'systems where work happens');
    await page.getByRole('toolbar', { name: 'Annotate selection' }).getByRole('button', { name: 'Comment' }).click();
    const composer = page.getByRole('dialog', { name: 'New note' });
    await composer.getByRole('textbox').fill(noteText);
    await composer.getByRole('button', { name: 'Post note' }).click();
    await expect(page.locator('mark[data-annotation-mark]', { hasText: 'systems where work happens' })).toBeVisible();

    // Reply from the sheet.
    await page.getByRole('button', { name: 'Open all notes' }).click();
    const sheet = page.locator('[data-slot=sheet-content]');
    const card = sheet.locator('div.corner-box-corners', { hasText: noteText }).first();
    await card.getByRole('button', { name: 'Reply', exact: true }).click();
    await card.getByRole('textbox').fill(replyText);
    await card.getByRole('button', { name: 'Post reply' }).click();

    await expect(card.locator('[data-reply-id]', { hasText: replyText })).toBeVisible();
    await expect(card.getByRole('button', { name: /1 reply/ })).toBeVisible();

    // Reply to the reply — flattened with an @mention.
    const replyRow = card.locator('[data-reply-id]', { hasText: replyText });
    await replyRow.getByRole('button', { name: /Reply to/ }).click();
    const nestedText = `Nested ${Date.now()}`;
    await card.getByRole('textbox').fill(nestedText);
    await card.getByRole('button', { name: 'Post reply' }).click();
    const nestedRow = card.locator('[data-reply-id]', { hasText: nestedText }).last();
    await expect(nestedRow).toBeVisible();
    await expect(nestedRow.getByText(/@E2E replier/)).toBeVisible();

    // Deleting the root removes the whole thread.
    await card.getByRole('button', { name: 'Delete note' }).first().click();
    await card.getByRole('button', { name: 'Confirm' }).click();
    await expect(sheet.getByText(noteText)).toBeHidden();
    await expect(sheet.getByText(replyText)).toBeHidden();
  });

  test('signed-out bookmarks page asks for sign-in', async ({ page }) => {
    await page.goto('/bookmarks');
    await expect(page.getByText('Sign in required', { exact: false })).toBeVisible();
  });
});
