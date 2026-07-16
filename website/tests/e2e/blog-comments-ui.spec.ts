import { expect, test } from '@playwright/test';

const POST_PATH = '/blog/agent-native-data-workflows';
const ROOT_BODY =
  'A durable note should stay readable after the pointer leaves the highlighted passage.';
const FIRST_REPLY = 'The first reply responds directly to the root and keeps its own actions.';
const NESTED_REPLY =
  'This second-level reply is visibly nested beneath its parent instead of flattened.';

function mockedComments() {
  const createdAt = new Date().toISOString();
  return {
    comments: [
      {
        id: 'root-note',
        body: ROOT_BODY,
        anchor: {
          sectionIndex: 0,
          blockType: 'body',
          blockIndex: 0,
          exact: 'LLM agents are good at planning across tools',
          prefix: '',
          suffix: ', but they',
          startOffset: 0,
        },
        parentId: null,
        createdAt,
        author: { name: 'Karthik Sivadas', image: null },
        mine: true,
      },
      {
        id: 'reply-one',
        body: FIRST_REPLY,
        anchor: null,
        parentId: 'root-note',
        createdAt,
        author: { name: 'Priya Nair', image: null },
        mine: false,
      },
      {
        id: 'reply-two',
        body: NESTED_REPLY,
        anchor: null,
        parentId: 'reply-one',
        createdAt,
        author: { name: 'Sam Lee', image: null },
        mine: false,
      },
    ],
    viewer: { admin: false },
  };
}

test.beforeEach(async ({ page }) => {
  await page.route(/\/api\/comments\?slug=agent-native-data-workflows$/, (route) =>
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(mockedComments()),
    }),
  );
});

test('pins a complete note and keeps it open after pointer movement', async ({ page }) => {
  await page.goto(POST_PATH);
  const mark = page.locator('mark[data-annotation-mark]').first();
  await expect(mark).toBeVisible();

  await mark.click();
  await page.mouse.move(8, 8);

  const pinned = page.locator('[data-note-preview="pinned"]');
  await expect(pinned).toBeVisible();
  await expect(pinned.getByText(ROOT_BODY, { exact: true })).toBeVisible();
  await expect(pinned.getByRole('button', { name: 'Open thread' })).toBeVisible();
});

test('renders replies beneath their direct parents and keeps messages visible on mobile', async ({
  page,
}) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(POST_PATH);
  await page.getByRole('button', { name: /Open notes/ }).click();

  const sheet = page.locator('[data-slot="sheet-content"]');
  await sheet.getByRole('button', { name: /2 replies/ }).click();

  const firstReply = sheet.locator('[data-reply-id="reply-one"]');
  const nestedReply = firstReply.locator('[data-reply-id="reply-two"]');
  await expect(firstReply).toHaveAttribute('data-thread-depth', '1');
  await expect(nestedReply).toHaveAttribute('data-thread-depth', '2');
  await expect(firstReply.getByText(FIRST_REPLY, { exact: true })).toBeVisible();
  await expect(nestedReply.getByText(NESTED_REPLY, { exact: true })).toBeVisible();

  const sheetBox = await sheet.boundingBox();
  expect(sheetBox).not.toBeNull();
  expect(sheetBox!.width).toBeGreaterThanOrEqual(370);
});

test('restores an optimistically removed note when delete loses the network', async ({ page }) => {
  await page.route(/\/api\/comments\/root-note$/, (route) => route.abort('failed'));
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(POST_PATH);
  await page.getByRole('button', { name: /Open notes/ }).click();

  const sheet = page.locator('[data-slot="sheet-content"]');
  await sheet.getByRole('button', { name: 'Delete note' }).click();
  await sheet.getByRole('button', { name: 'Confirm' }).click();

  await expect(sheet.getByText(ROOT_BODY, { exact: true })).toBeVisible();
});
