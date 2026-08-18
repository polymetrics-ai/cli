# Discussion log — #3859

The launch brief and issue resolve the material product choices without an
interactive pause:

1. Deliver all six canonical destination modes, retaining the public
   PostgreSQL `write=false` capability fence.
2. Treat a durable acknowledgement as a full target transaction plus durable
   delivery evidence, never an individual statement or a successful return.
3. Require a descriptor-selected, closed apply request. There is no raw SQL,
   arbitrary relation expression, HTTP request, or shell execution path.
4. Use the source ordering tuple as the only ordering authority. Extraction
   clock/order is not an admissible substitute.
5. Preserve physical absence; only an explicit tombstone may remove/close a
   target record.

The one potentially overlapping dependency is #3858. It remains intentionally
untouched: this target side accepts its sealed page boundary and returns only
durable acknowledgement evidence.
