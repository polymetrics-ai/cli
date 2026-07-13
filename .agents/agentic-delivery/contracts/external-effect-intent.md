# External-effect intent

GSD may request only typed issue create/update/close, PR create/update/close, comment, branch, or push
effects. An intent is not authority. Shepherd requires a matching grant, generation, fence, target,
payload hash, and idempotency key. Merge and branch deletion are not valid intent kinds.

