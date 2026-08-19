# Final classification ledger

The following buckets are disjoint and cover command IDs `1` through `146` exactly once.

- `certified` (29): `11`, `12`, `15`, `19`, `20`, `63`, `65`, `66`, `67`, `68`, `69`, `70`, `71`, `73`, `74`, `75`, `79`, `80`, `83`, `88`, `92`, `93`, `94`, `95`, `102`, `103`, `104`, `105`, `136`.
- `no_object` (1): `8`. The parent cache collection was empty and GitHub exposes no REST create-cache operation; creating one requires an Actions workflow execution, which crosses the real-money boundary.
- `wrong_credential` (0): none. Recoverable credential failures were retried through the measured credential route before classification.
- `entitlement` (22): `10`, `32`, `42`, `44`, `56`, `58`, `59`, `60`, `61`, `76`, `116`, `117`, `118`, `119`, `124`, `125`, `126`, `127`, `132`, `133`, `134`, `144`.
- `not_implemented` (1): `146`.
- `product_defect` (88): `1`, `2`, `3`, `4`, `5`, `6`, `7`, `9`, `13`, `14`, `16`, `17`, `18`, `21`, `22`, `23`, `24`, `25`, `26`, `27`, `28`, `29`, `30`, `31`, `33`, `34`, `35`, `36`, `37`, `38`, `39`, `40`, `41`, `43`, `45`, `46`, `47`, `48`, `49`, `50`, `51`, `52`, `53`, `54`, `55`, `57`, `62`, `64`, `72`, `77`, `78`, `81`, `82`, `84`, `85`, `86`, `87`, `89`, `90`, `91`, `96`, `97`, `98`, `99`, `100`, `101`, `106`, `107`, `108`, `109`, `110`, `111`, `112`, `113`, `114`, `115`, `128`, `129`, `130`, `131`, `135`, `137`, `138`, `139`, `140`, `141`, `143`, `145`.
- `escape_needs_captain` (5): `120`, `121`, `122`, `123`, `142`.

Total: `29 + 1 + 0 + 22 + 1 + 88 + 5 = 146`.

Commands `120`–`123` issued no provider request because they would create **public visibility under the org's name**. Command `142` issued no provider request because the hosted agent's autonomous token and Actions consumption made the per-operation cost genuinely unknowable. These are deliberate captain-bound outcomes, not missing attempts.
