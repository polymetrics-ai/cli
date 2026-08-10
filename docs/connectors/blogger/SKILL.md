---
name: pm-blogger
description: Blogger connector knowledge and safe action guide.
---

# pm-blogger

## Purpose

Reads Blogger (Google Blogger API v3) blogs, posts, pages, comments, and page-view counts using an OAuth 2.0 refresh-token grant. Read-only.

## Icon

- id: simple-icons-blogger
- asset: icons/simple-icons/blogger.svg
- title: Blogger
- simple_icon_slug: blogger
- simple_icon_hex: FF5722
- source: simple-icons
- license: CC0-1.0
- review_status: cc0_with_trademark_caveat
- review_url: https://simpleicons.org/?q=Blogger
- match: exact-name-or-slug
- matched_by: blogger

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- blog_id
- page_size
- token_url
- client_id (secret)
- client_refresh_token (secret)
- client_secret (secret)

## ETL Streams

- blogs:
  - primary key: id
  - cursor: updated
  - fields: description(string), id(string), kind(string), name(string), pages_total(integer), posts_total(integer), published(string), updated(string), url(string)
- posts:
  - primary key: id
  - cursor: updated
  - fields: author_display_name(string), author_id(string), blog_id(string), content(string), id(string), kind(string), published(string), replies_total(integer), status(string), title(string), updated(string), url(string)
- pages:
  - primary key: id
  - cursor: updated
  - fields: author_display_name(string), author_id(string), blog_id(string), content(string), id(string), kind(string), published(string), status(string), title(string), updated(string), url(string)
- comments:
  - primary key: id
  - cursor: updated
  - fields: author_display_name(string), author_id(string), blog_id(string), content(string), id(string), kind(string), post_id(string), published(string), status(string), updated(string)
- pageviews:
  - primary key: blog_id, time_range
  - fields: blog_id(string), count(string), time_range(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Security

- read risk: external Blogger API read of blog/post/page/comment metadata and page-view counts
- approval: none; read-only, no reverse-ETL write surface
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run Blogger's declared streams and reverse-ETL actions.
- Usage: pm blogger <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - api delete blogs blogid posts postid - Documented DELETE /blogs/{blogId}/posts/{postId} (not implemented) [intent=direct_write availability=not_implemented operation=blogger.delete.blogs-blogid-posts-postid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v3 blogs blogid pages pageid - Documented DELETE /v3/blogs/{blogId}/pages/{pageId} (not implemented) [intent=direct_write availability=not_implemented operation=blogger.delete.v3-blogs-blogid-pages-pageid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v3 blogs blogid posts postid - Documented DELETE /v3/blogs/{blogId}/posts/{postId} (not implemented) [intent=direct_write availability=not_implemented operation=blogger.delete.v3-blogs-blogid-posts-postid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete v3 blogs blogid posts postid comments commentid - Documented DELETE /v3/blogs/{blogId}/posts/{postId}/comments/{commentId} (not implemented) [intent=direct_write availability=not_implemented operation=blogger.delete.v3-blogs-blogid-posts-postid-comments-commentid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get blogs blogid comments commentid - Documented GET /blogs/{blogId}/comments/{commentId} (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.blogs-blogid-comments-commentid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get blogs blogid posts postid - Documented GET /blogs/{blogId}/posts/{postId} (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.blogs-blogid-posts-postid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get blogs byurl - Documented GET /blogs/byurl (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.blogs-byurl]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid - Documented GET /v3/blogs/{blogId} (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid comments - Documented GET /v3/blogs/{blogId}/comments (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-comments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid pages - Documented GET /v3/blogs/{blogId}/pages (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-pages]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid pages pageid - Documented GET /v3/blogs/{blogId}/pages/{pageId} (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-pages-pageid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid pageviews - Documented GET /v3/blogs/{blogId}/pageviews (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-pageviews]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid posts - Documented GET /v3/blogs/{blogId}/posts (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-posts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid posts bypath - Documented GET /v3/blogs/{blogId}/posts/bypath (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-posts-bypath]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid posts postid - Documented GET /v3/blogs/{blogId}/posts/{postId} (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-posts-postid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid posts postid comments - Documented GET /v3/blogs/{blogId}/posts/{postId}/comments (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-posts-postid-comments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid posts postid comments commentid - Documented GET /v3/blogs/{blogId}/posts/{postId}/comments/{commentId} (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-posts-postid-comments-commentid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs blogid posts search - Documented GET /v3/blogs/{blogId}/posts/search (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-blogid-posts-search]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 blogs byurl - Documented GET /v3/blogs/byurl (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-blogs-byurl]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 users userid - Documented GET /v3/users/{userId} (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-users-userid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 users userid blogs - Documented GET /v3/users/{userId}/blogs (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-users-userid-blogs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 users userid blogs blogid - Documented GET /v3/users/{userId}/blogs/{blogId} (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-users-userid-blogs-blogid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 users userid blogs blogid posts - Documented GET /v3/users/{userId}/blogs/{blogId}/posts (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-users-userid-blogs-blogid-posts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get v3 users userid blogs blogid posts postid - Documented GET /v3/users/{userId}/blogs/{blogId}/posts/{postId} (not implemented) [intent=direct_read availability=not_implemented operation=blogger.get.v3-users-userid-blogs-blogid-posts-postid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api patch v3 blogs blogid pages pageid - Documented PATCH /v3/blogs/{blogId}/pages/{pageId} (not implemented) [intent=direct_write availability=not_implemented operation=blogger.patch.v3-blogs-blogid-pages-pageid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api patch v3 blogs blogid posts postid - Documented PATCH /v3/blogs/{blogId}/posts/{postId} (not implemented) [intent=direct_write availability=not_implemented operation=blogger.patch.v3-blogs-blogid-posts-postid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post blogs blogid comments commentid moderate-true - Documented POST /blogs/{blogId}/comments/{commentId}?moderate=true (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.blogs-blogid-comments-commentid-moderate-true]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post blogs blogid posts - Documented POST /blogs/{blogId}/posts (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.blogs-blogid-posts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 blogs blogid pages - Documented POST /v3/blogs/{blogId}/pages (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.v3-blogs-blogid-pages]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 blogs blogid pages pageid publish - Documented POST /v3/blogs/{blogId}/pages/{pageId}/publish (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.v3-blogs-blogid-pages-pageid-publish]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 blogs blogid pages pageid revert - Documented POST /v3/blogs/{blogId}/pages/{pageId}/revert (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.v3-blogs-blogid-pages-pageid-revert]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 blogs blogid posts - Documented POST /v3/blogs/{blogId}/posts (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.v3-blogs-blogid-posts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 blogs blogid posts postid comments commentid approve - Documented POST /v3/blogs/{blogId}/posts/{postId}/comments/{commentId}/approve (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.v3-blogs-blogid-posts-postid-comments-commentid-approve]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 blogs blogid posts postid comments commentid removecontent - Documented POST /v3/blogs/{blogId}/posts/{postId}/comments/{commentId}/removecontent (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.v3-blogs-blogid-posts-postid-comments-commentid-removecontent]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 blogs blogid posts postid comments commentid spam - Documented POST /v3/blogs/{blogId}/posts/{postId}/comments/{commentId}/spam (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.v3-blogs-blogid-posts-postid-comments-commentid-spam]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 blogs blogid posts postid publish - Documented POST /v3/blogs/{blogId}/posts/{postId}/publish (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.v3-blogs-blogid-posts-postid-publish]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post v3 blogs blogid posts postid revert - Documented POST /v3/blogs/{blogId}/posts/{postId}/revert (not implemented) [intent=direct_write availability=not_implemented operation=blogger.post.v3-blogs-blogid-posts-postid-revert]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put blogs blogid posts postid - Documented PUT /blogs/{blogId}/posts/{postId} (not implemented) [intent=direct_write availability=not_implemented operation=blogger.put.blogs-blogid-posts-postid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 blogs blogid pages pageid - Documented PUT /v3/blogs/{blogId}/pages/{pageId} (not implemented) [intent=direct_write availability=not_implemented operation=blogger.put.v3-blogs-blogid-pages-pageid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put v3 blogs blogid posts postid - Documented PUT /v3/blogs/{blogId}/posts/{postId} (not implemented) [intent=direct_write availability=not_implemented operation=blogger.put.v3-blogs-blogid-posts-postid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - blogs list - Run the blogs ETL stream [intent=etl availability=implemented stream=blogs]; notes: discrepancy=present-in-surface-absent-from-artifact
  - comments list - Run the comments ETL stream [intent=etl availability=implemented stream=comments]; notes: discrepancy=present-in-surface-absent-from-artifact
  - pages list - Run the pages ETL stream [intent=etl availability=implemented stream=pages]; notes: discrepancy=present-in-surface-absent-from-artifact
  - pageviews list - Run the pageviews ETL stream [intent=etl availability=implemented stream=pageviews]; notes: discrepancy=present-in-surface-absent-from-artifact
  - posts list - Run the posts ETL stream [intent=etl availability=implemented stream=posts]; notes: discrepancy=present-in-surface-absent-from-artifact

## Commands

### Inspect as a manual

```bash
pm connectors inspect blogger
```

### Inspect as structured JSON

```bash
pm connectors inspect blogger --json
```

## Agent Rules

- Run pm connectors inspect blogger before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
