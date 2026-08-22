# TDD Ledger: Foundation public-output repair r1

| Finding | Red evidence to add | Green behavior | Edge and regression evidence |
| --- | --- | --- | --- |
| FND-B10 | Name-shaped SQS/provider output is wrongly redacted. | Only configured material and its concrete encodings are masked. | `id`, `occurrence_id`, token-shaped values, headers, and key names remain byte-for-byte intact. |
| FND-B11 | Public cursor/receipt serialization contains configured material. | Every public cursor and receipt projection masks it while retaining ordinary identifiers. | JSON escaped and printable/encoded values cannot bypass masking. |
| FND-B12 | Invalid GitHub App restrictions continue to an authenticated request. | Invalid syntax/semantics returns a validation error before I/O. | Valid restriction succeeds; malformed empty/duplicate/out-of-range forms retain a zero request counter. |
| FND-B13 | Non-JSON public diagnostics expose configured material or erase ordinary text. | Printable text forms mask only configured material/representations. | Exact harmless words, provider IDs, and occurrence IDs remain unchanged. |
| FND-B14 | Binary download binds undeclared/unsafe parameters. | The declared parameter authority gate rejects before I/O. | Declared parameters make one exact request; unknown, cross-operation, invalid, and unsafe values make none. |
| FND-W02 | Status execution admits bindings outside the declaration. | Status uses the same authority gate as operation requests. | Valid declared parameters preserve request fidelity and invalid bindings make no request. |

## Actual evidence

Pending red-first execution. Each command and result is appended as its slice becomes green.
