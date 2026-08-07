# workday-rest — TDD ledger

Program `cli-top50-fixed-schema-sweep-r1` · branch `fm/cli-top50-sweep-resume2-r1` · landing order #2
(largest-first), behind github.

Every entry records the failure **verbatim** as `go test` printed it. Nothing here is paraphrased,
and no assertion in this file has been weakened to reach green.

---

## Cycle 1 — Red (slice 1, previous worker): documented surface is not enumerated

**Test:** `cmd/connectorgen/workday_rest_documented_surface_test.go`
`TestWorkdayRESTDocumentedSurfaceIsComplete`

**Red:** observed against the real shipped bundle (4 rows), constant at 916.

```
=== RUN   TestWorkdayRESTDocumentedSurfaceIsComplete
    workday_rest_documented_surface_test.go:118: operation_ledger_version = 0, want 1
    workday_rest_documented_surface_test.go:204: 1 legacy excluded row(s) remain, want 0
    workday_rest_documented_surface_test.go:207: documented endpoints = 0, want 916 (920 raw rows across 52 service specs, minus 4 documented twice; 4 legacy /ccx/ row(s) counted apart)
    workday_rest_documented_surface_test.go:216: covered(3)+blocked(0) = 3, want 4 - every row needs a disposition, legacy included
    workday_rest_documented_surface_test.go:220: byMethod = map[], want map[DELETE:32 GET:654 PATCH:58 POST:153 PUT:19]
    workday_rest_documented_surface_test.go:227: expected the collapsed custom-object row "POST /customObject/v2/customObjects/{customObjectAlias}" exactly once
    workday_rest_documented_surface_test.go:241: expected "GET /absenceManagement/v5/workers" - the shipped bundle enumerated only the three legacy HCM read streams
--- FAIL: TestWorkdayRESTDocumentedSurfaceIsComplete (0.00s)
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.503s
FAIL
```

**Green:** not yet. Slices 2–4 author the surface.

---

## Cycle 2 — Red re-observed at the corrected count (slice 2, this worker): 916 → **907**

### Why the constant moved, and why that is a tightening rather than a weakening

The slice-1 derivation deduped on the resolved `(method, base+path)` pair. That caught the four
custom-object rows published by two service modules (920 → 916) and it **never looked for a `?`**.

Nine of the 916 are query-string variants of an endpoint already counted. The provider publishes
each as its own Swagger path key with the query string baked in, and **every one has its base-path
sibling documented separately**, so collapsing them loses no endpoint:

| Method | Variant | Collapses into |
| --- | --- | --- |
| GET | `/accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments?type=viewContent` | same path |
| GET | `/accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments/{subresourceID}?type=viewContent` | same path |
| GET | `/procurement/v5/requisitions/{ID}/attachments?type=getFileContent` | same path |
| GET | `/procurement/v5/requisitions/{ID}/attachments/{subresourceID}?type=getFileContent` | same path |
| GET | `/recruiting/v4/prospects/{ID}/resumeAttachments?type=viewFile` | same path |
| GET | `/recruiting/v4/prospects/{ID}/resumeAttachments/{subresourceID}?type=viewFile` | same path |
| PATCH | `/staffing/v7/workers/{ID}/checkInTopics/{subresourceID}?type=archive` | same path |
| PATCH | `/staffing/v7/workers/{ID}/checkIns/{subresourceID}?type=archive` | same path |
| POST | `/api/common/v1/workers/{ID}/businessTitleChanges?type=me` | same path |

Two independent facts settle them as variants rather than operations:

- **Seven carry an empty summary.** The provider documents them as an addendum to the base row.
  Procurement's base row says so outright — *"Retrieves the metadata **or the attachment content**
  of the specified requisition"* — one endpoint, two modes.
- **The two staffing PATCHes carry a summary describing a behaviour** of the base endpoint
  (*"…to archived or un-archived"*), which is sweep finding 23's shape exactly.

**The assertion got stricter, not looser.** The corrected test additionally requires that each
collapsed variant's **base endpoint is present**, so a double-count can never be "fixed" into a
missing operation — the failure mode that would otherwise hide behind a smaller number.

**The artifact was re-fetched, not trusted.** Manifest HTTP 200 at **617,538 bytes**, byte-identical
to slice 1; all 52 specs re-fetched and the derivation reproduced independently at 920 raw /
`GET 655 · POST 154 · PATCH 58 · DELETE 33 · PUT 20` / 916 after cross-service dedup. Slice 1 is
confirmed exactly to that point.

**Red:**

```
--- FAIL: TestWorkdayRESTDocumentedSurfaceIsComplete (0.00s)
    workday_rest_documented_surface_test.go:158: operation_ledger_version = 0, want 1
    workday_rest_documented_surface_test.go:244: 1 legacy excluded row(s) remain, want 0
    workday_rest_documented_surface_test.go:247: documented endpoints = 0, want 907 (920 raw rows across 52 service specs, minus 4 published by two service modules, minus 9 query-string variants of an endpoint already counted; 4 legacy /ccx/ row(s) counted apart)
    workday_rest_documented_surface_test.go:258: covered(3)+blocked(0) = 3, want 4 — every row needs a disposition, legacy included
    workday_rest_documented_surface_test.go:262: byMethod = map[], want map[DELETE:32 GET:648 PATCH:56 POST:152 PUT:19]
    workday_rest_documented_surface_test.go:269: expected the collapsed custom-object row "DELETE /customObject/v2/customObjects/{customObjectAlias}/{customObjectID}" exactly once
    workday_rest_documented_surface_test.go:269: expected the collapsed custom-object row "GET /customObject/v2/customObjects/{customObjectAlias}/{customObjectID}" exactly once
    workday_rest_documented_surface_test.go:269: expected the collapsed custom-object row "POST /customObject/v2/customObjects/{customObjectAlias}" exactly once
    workday_rest_documented_surface_test.go:269: expected the collapsed custom-object row "PUT /customObject/v2/customObjects/{customObjectAlias}/{customObjectID}" exactly once
    workday_rest_documented_surface_test.go:284: expected "GET /accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments/{subresourceID}" — the endpoint that "GET /accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments/{subresourceID}?type=viewContent" collapses into
    workday_rest_documented_surface_test.go:284: expected "GET /accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments" — the endpoint that "GET /accountsPayable/v1/supplierInvoiceRequests/{ID}/attachments?type=viewContent" collapses into
    workday_rest_documented_surface_test.go:284: expected "GET /procurement/v5/requisitions/{ID}/attachments/{subresourceID}" — the endpoint that "GET /procurement/v5/requisitions/{ID}/attachments/{subresourceID}?type=getFileContent" collapses into
    workday_rest_documented_surface_test.go:284: expected "GET /procurement/v5/requisitions/{ID}/attachments" — the endpoint that "GET /procurement/v5/requisitions/{ID}/attachments?type=getFileContent" collapses into
    workday_rest_documented_surface_test.go:284: expected "GET /recruiting/v4/prospects/{ID}/resumeAttachments/{subresourceID}" — the endpoint that "GET /recruiting/v4/prospects/{ID}/resumeAttachments/{subresourceID}?type=viewFile" collapses into
    workday_rest_documented_surface_test.go:284: expected "GET /recruiting/v4/prospects/{ID}/resumeAttachments" — the endpoint that "GET /recruiting/v4/prospects/{ID}/resumeAttachments?type=viewFile" collapses into
    workday_rest_documented_surface_test.go:284: expected "PATCH /staffing/v7/workers/{ID}/checkInTopics/{subresourceID}" — the endpoint that "PATCH /staffing/v7/workers/{ID}/checkInTopics/{subresourceID}?type=archive" collapses into
    workday_rest_documented_surface_test.go:284: expected "PATCH /staffing/v7/workers/{ID}/checkIns/{subresourceID}" — the endpoint that "PATCH /staffing/v7/workers/{ID}/checkIns/{subresourceID}?type=archive" collapses into
    workday_rest_documented_surface_test.go:284: expected "POST /api/common/v1/workers/{ID}/businessTitleChanges" — the endpoint that "POST /api/common/v1/workers/{ID}/businessTitleChanges?type=me" collapses into
    workday_rest_documented_surface_test.go:296: expected "GET /absenceManagement/v5/workers" — the shipped bundle enumerated only the three legacy HCM read streams
    workday_rest_documented_surface_test.go:296: expected "GET /accountsPayable/v1/supplierInvoiceRequests" — the shipped bundle enumerated only the three legacy HCM read streams
    workday_rest_documented_surface_test.go:296: expected "GET /api/prismAnalytics/v3/{tenant}/tables" — the shipped bundle enumerated only the three legacy HCM read streams
    workday_rest_documented_surface_test.go:296: expected "POST /customObject/v2/customObjects/{customObjectAlias}" — the shipped bundle enumerated only the three legacy HCM read streams
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.540s
FAIL
```

**Green:** not yet — slices 2–4 author the surface against this corrected assertion.

**Shared gates at this commit:** green. `connectorgen validate` 551/0, `surface-sync --check` clean,
runtime preflight passing. Only workday-rest's own surface test is red, by design, exactly as
github's was mid-delivery.
