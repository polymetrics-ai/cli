# Security Trace

- Direct reads only support GET.
- Absolute URLs and mutation methods are rejected before network I/O.
- Missing path variables and traversal paths fail before network I/O.
- Response size is bounded and HTTP errors are redacted.
- No generic raw API, HTTP write, SQL, shell, or reverse-ETL bypass was added.
