# Overview

Fanout Bad Connector is a synthetic connector used as a connectorgen validate
seeded-invalid case (fan_out.ids_from.request.path referencing an undeclared
spec key).

## Auth setup

None; public test fixture.

## Streams notes

`tasks` fans out over `/projects`.

## Write actions & risks

None.

## Known limits

None; this is test fixture data.
