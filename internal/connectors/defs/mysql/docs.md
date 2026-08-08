# Overview

Reads MySQL tables through the MySQL wire protocol. It discovers tables and columns dynamically,
then performs bounded full and cursor-incremental reads. It is a read-only source registered through
the native MySQL factory.

## Auth setup

Configure a bare `host`, `database`, and `username`. `port` defaults to 3306. `password` is a
secret field and is never logged. Do not put credentials in a host or URL-shaped value.

`sslmode` chooses transport security:

| `sslmode` | Encrypts | Falls back to plaintext | Verifies chain | Verifies server name |
| --- | --- | --- | --- | --- |
| `disabled` | no | n/a | no | no |
| `preferred` (default) | when offered | yes, only when the server advertises no TLS | no | no |
| `required` | yes | never | no | no |
| `verify-ca` | yes | never | yes | no |
| `verify-identity` | yes | never | yes | yes |

`preferred` is the only mode that may fall back. `required` and both verifying modes fail closed
when the server offers no TLS. `sslrootcert` names an absolute PEM CA path for the verifying modes
and defaults to the host trust store; `sslservername` overrides the verified name under
`verify-identity`, for example when connecting to an IP address. Compatibility spellings
(`disable`/`allow`/`prefer`/`require`/`verify-full` and `verify_ca`/`verify_identity`) are accepted;
use the canonical modes in the table for a portable policy across SQL connectors.

## Streams notes

The catalog exposes every base table in the configured database as `database.table`. Complete
read paging requires a single-column primary key. Without `cursor_field`, a full snapshot is
unfiltered and pages by that primary key. Set `cursor_field` only to a non-null single-column
primary or unique key for incremental reads; pages order by `(cursor_field, primary_key)` and
resume strictly after a present stored cursor, including an empty or whitespace text value.
`page_size` and `read_limit` bound every read. Textual and temporal wire values are emitted as
strings, while binary values are copied before emission.

## Write actions & risks

This connector is read-only. It issues discovery and read queries; it does not create, alter, or
delete database data.

## Known limits

- MySQL 8.4.11 is covered by an opt-in container integration test.
- Tables and columns whose identifiers this connector cannot safely quote are omitted from the
  catalog rather than advertised, because a Read against them would always fail.
- Client certificate authentication is outside this connector today.
