# Test Plan

## Red Tests Before Code

- `TestReadGraphQLBodySendsFixedDocumentAndVariables`
  - proves a POST stream with `graphql` metadata sends `query`, `operationName`, and resolved
    `variables` in the JSON body.
- `TestReadGraphQLErrorsFailClosed`
  - proves a top-level non-empty GraphQL `errors` array is returned as an engine error before record
    extraction.
- `TestWriteGraphQLBodyUsesFixedDocumentAndDeclaredVariables`
  - proves `body_type: graphql` sends the static bundle document and declared variables.
- `TestWriteGraphQLBodyIgnoresRecordQueryField`
  - proves record input cannot override the static GraphQL document.

## Green Tests After Code

- Run the same focused tests.
- Run the full engine package tests.
- Validate the GitHub connector bundle still loads.

## Regression Boundary

Existing JSON, form, and none body writes must keep their current behavior. REST stream requests with
no body must still send no body.
