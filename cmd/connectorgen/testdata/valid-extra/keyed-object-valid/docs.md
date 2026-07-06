# Overview

Keyed Object Valid is a control connectorgen validate corpus case (S4 engine mini-wave item 3):
its `streams.json` `products` stream declares a well-formed `records.keyed_object`/`key_field`
block (appfigures' `products/mine` shape: a JSON object keyed by product id rather than an array).

## Auth setup

None; public test fixture.

## Streams notes

`products` flattens the keyed object at `products`, stamping each key onto `product_id`.

## Write actions & risks

None; read-only test fixture.

## Known limits

None; this is test fixture data.
