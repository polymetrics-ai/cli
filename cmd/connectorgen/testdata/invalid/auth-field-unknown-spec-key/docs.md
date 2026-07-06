# Overview

Auth Field Unknown Spec Key is a deliberately invalid connectorgen validate corpus case: its
`streams.json` `base.auth`'s `basic` mode templates `username` against an undeclared spec key
(`config.nope_username`), which `engine.ResolveCheckAuthSpec` must reject.

## Auth setup

Basic auth, deliberately misconfigured for this test case.

## Streams notes

`widgets` is a single stream, no pagination.

## Write actions & risks

`update_widget` is a low-risk PATCH.

## Known limits

None; this is test fixture data.
