# Original foundation proof inputs

These exact historical bytes match the immutable proof document and its original
selected execution input pins. They are used only in disposable proof-reader
fixtures after CP14 changed live registry metadata and Atlas ownership. They do
not establish current-source execution or update the reviewed proof allowlist.
The test helper checks hashes before copying either file. Current-corpus readers
continue to observe actual changed inputs as stale.

[
  {
    "path": "docs/connector-canon/foundations/catalog.json",
    "sha256": "77bdc26f357bc57490f98cffb7df2ca427f4041429f38b225df63688ea9a6b92",
    "bytes": 182690,
    "source_commit": "172cf0cce679b46e81671de03a12e63f019dfd6c",
    "fixture": "77bdc26f357bc57490f98cffb7df2ca427f4041429f38b225df63688ea9a6b92.artifact"
  },
  {
    "path": "internal/connectors/connectors.go",
    "sha256": "c3d447c603628929911a6d2cf6716285da8597eb3bb8aee9715ba3a533441362",
    "bytes": 138075,
    "source_commit": "172cf0cce679b46e81671de03a12e63f019dfd6c",
    "fixture": "c3d447c603628929911a6d2cf6716285da8597eb3bb8aee9715ba3a533441362.artifact"
  }
]

## Complete frozen input closure (CI-175-01)

`inputs.json` records all231 original execution inputs plus the Atlas. Every file was recovered from the recorded `172cf0cce679b46e81671de03a12e63f019dfd6c` commit and independently checked against the unchanged proof document byte count and SHA256 before retention. The proof-reader fixture verifies every input before copying it into its disposable root. Captures and outputs remain their original separately pinned records. This replaces the former two-file exception and prevents newer live source edits from corrupting a historical positive fixture. These files are not compiled or executed as product code and cannot establish current-corpus proof.
