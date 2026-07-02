# Overview

Goodconn is a synthetic connector used as the connectorgen validate control case.

## Auth setup

Provide a bearer token via the `token` secret.

## Streams notes

`widgets` is incremental on `updated_at`.

## Write actions & risks

`update_widget` is a low-risk PATCH.

## Known limits

None; this is test fixture data.
