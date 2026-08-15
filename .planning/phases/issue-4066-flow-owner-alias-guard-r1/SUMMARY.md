# #4066 Summary

The final #3897 correction makes omitted-connection flow queries fail closed
when DuckDB aliases or ASCII-case-equivalent real tables could otherwise bypass
the warehouse resolver's typed ambiguity. It leaves generic SQL access to real
tables intact and does not reserve or rename user table names.

The correction is recorded as #3897 correction 5/5. It shares existing branch
`feat/3897-flow-connection-scope-nm` and existing draft PR #4060, whose base
must remain `feat/3988-github-certification`.
