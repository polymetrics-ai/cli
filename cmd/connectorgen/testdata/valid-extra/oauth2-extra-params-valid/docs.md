# Overview

OAuth2 Extra Params Valid is a control connectorgen validate corpus case (S4 engine mini-wave item
4): its `streams.json` `base.auth` declares a well-formed `oauth2_client_credentials` candidate with
an `extra_params.audience` derived from `config.base_url` — auth0's own audience-derivation
convention.

## Auth setup

OAuth2 client-credentials with an audience extra param derived from `base_url`.

## Streams notes

`widgets` is a single stream, no pagination.

## Write actions & risks

None; read-only test fixture.

## Known limits

None; this is test fixture data.
