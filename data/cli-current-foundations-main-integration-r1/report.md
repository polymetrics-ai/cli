# Current Foundations Main Integration r1 — Intake Report

**Base:** `main` / `e62ae21d428f0d27225f9bff564dc2cd797f6b65`.

The base already contains the non-empty credential-input foundation from #4310 and the slim production source-lock embed from #4309. The component pull requests remain open and unmodified.

Remote pull-heads re-read through `gh-axi` and verified against `origin` are recorded in `input-manifest.json`. Eligible inputs are #4312 (`19a32bd0bc08faf217be8f45b39841b5ff589a92`), #4311 (`3c768cade6703426afd2272fbc01bfd60583e04f`), and #4308 (`fe5b8e18788538c4fcce34969da7ff88a7fa66d6`). Firstmate supplied #4308's successful Verify run `32354323284` and its real-provider qualification for that exact SHA. #4305 and #4303 remain held until Firstmate supplies the final published no-mistakes heads; the visible #4304 head has failing checks and is not an input.

#4312 and #4311 are composed through preserved merge commits. The #4308 probe exposed a genuine overlap: #4311's typed declaration-owned header retention uses the ordinary response path, while #4308's status-only contract must retain a final non-2xx result after retry. The production-shaped red test now reproduces the loss as an `http 404` error before final metadata is returned. Resolution must keep the typed header path while using the closed status-only response method; it must not alter ordinary binary/text GET error semantics.

The remaining integration boundary is the outstanding #4305 and #4303 handoff. No older or unpushed head has been substituted.
