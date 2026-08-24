# Outreach Custom Objects documentation audit

Captured 2026-08-24 from Outreach's public rendered documentation. This is a
separate, immutable citation for the six fixed Custom Objects operations that
the provider documents outside its v2 OpenAPI artifact. It does not alter the
legacy OpenAPI pin, rewrite a drifted document, or claim a per-account custom
object schema: those schemas remain dynamic.

| Provider documentation | Retrieval result | Pinned response identity | Document location | Fixed operations evidenced |
| --- | --- | --- | --- | --- |
| `https://developers.outreach.io/api/custom-objects` | HTTP 200; `text/html; charset=UTF-8`; no redirect | 422,602 bytes; SHA-256 `702ebe938afaf106accaf1378cba916d6780d41a6a3d88b9f2a6de6cf846cd9d` | **Custom Objects Via API** → **Accessing Custom Object Endpoints** and **CRUD operations on custom objects** | `GET /api/v2/customObjects/{objectName}`, `GET /api/v2/customObjects/{objectName}/{id}`, `POST /api/v2/customObjects/{objectName}`, `PATCH /api/v2/customObjects/{objectName}/{id}`, `DELETE /api/v2/customObjects/{objectName}/{id}`, and `GET /api/v2/schema` |

The page also documents dynamic object-type names, filtering, related-resource
inclusion, and sparse fieldsets. Those facts support the existing generic
route templates only; they do not make an account-specific object request or
response schema closed. The citation will become the second v3
`rendered_reference` source document when the immutable Outreach OpenAPI
artifact and shared v3 operation-evidence projection are both available.

The 2026-08-19 `api_surface.json` supplement previously cited the same page
with an unretained digest. This audit is the current real document capture and
is not a re-pin of that historic supplement.
