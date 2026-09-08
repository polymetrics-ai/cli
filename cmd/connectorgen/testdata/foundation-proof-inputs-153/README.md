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
