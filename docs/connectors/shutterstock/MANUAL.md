# pm connectors inspect shutterstock

```text
NAME
  pm connectors inspect shutterstock - Shutterstock connector manual

SYNOPSIS
  pm connectors inspect shutterstock
  pm connectors inspect shutterstock --json
  pm credentials add <name> --connector shutterstock [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Shutterstock media, collection, license, editorial, catalog, contributor, and subscription metadata; writes collection/lightbox metadata through safe collection endpoints.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  audio_collection_ids
  audio_ids
  base_url
  category
  contributor_collection_ids
  contributor_id
  contributor_ids
  editorial_ids
  editorial_image_ids
  editorial_image_livefeed_ids
  editorial_livefeed_ids
  editorial_video_ids
  image_collection_ids
  image_ids
  orientation
  query
  sfx_ids
  sort
  video_collection_ids
  video_ids
  visual_asset_ids
  access_token (secret)

ETL STREAMS
  images:
    primary key: id
    cursor: updated_at
    fields: description(string), id(string), media_type(string), updated_at(string)
  videos:
    primary key: id
    cursor: updated_at
    fields: description(string), id(string), media_type(string), updated_at(string)
  audio:
    primary key: id
    cursor: updated_at
    fields: description(string), id(string), media_type(string), updated_at(string)
  list_similar_images:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_image_recommendations:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_updated_images:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_image_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_image:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_image_license_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_image_collection_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_image_collection:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_image_collection_items:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  find_similar_videos:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_updated_videos:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_video_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_video:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_video_license_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_video_collection_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_video_collection:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_video_collection_items:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_track_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_track:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_track_license_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_track_collection_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_track_collection:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_track_collection_items:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  search_sfx:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_sfx:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_sfx_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_sfx_license_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  search_editorial_images:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_updated_editorial_images:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_image:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_image_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_image_license_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_image_livefeeds:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_image_livefeed:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_image_livefeed_items:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_livefeeds:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_livefeed:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_livefeed_items:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  search_editorial:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_updated_editorial:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  search_editorial_videos:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_video:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_video_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_editorial_video_license_list:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_similar_cv_images:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_similar_cv_videos:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  search_catalog:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_collections:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_contributors:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_contributor:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_contributor_collections:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_contributor_collection:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_contributor_collection_items:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)
  get_user_subscriptions:
    primary key: id
    fields: asset_type(string), collection_id(string), contributor_id(string), created_at(string), created_time(string), description(string), id(string), items(array), license_id(string), media_type(string), name(string), subscription_id(string), updated_at(string), updated_time(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_image_collection:
    endpoint: POST /v2/images/collections
    required fields: name
    risk: external Shutterstock POST /v2/images/collections; approval required
  rename_image_collection:
    endpoint: POST /v2/images/collections/{{ record.image_collection_id }}
    required fields: image_collection_id, name
    risk: external Shutterstock POST /v2/images/collections/{{ record.image_collection_id }}; approval required
  delete_image_collection:
    endpoint: DELETE /v2/images/collections/{{ record.image_collection_id }}
    required fields: image_collection_id
    risk: destructive external Shutterstock DELETE /v2/images/collections/{{ record.image_collection_id }}; approval required
  add_image_collection_items:
    endpoint: POST /v2/images/collections/{{ record.image_collection_id }}/items
    required fields: image_collection_id, items
    risk: external Shutterstock POST /v2/images/collections/{{ record.image_collection_id }}/items; approval required
  create_video_collection:
    endpoint: POST /v2/videos/collections
    required fields: name
    risk: external Shutterstock POST /v2/videos/collections; approval required
  rename_video_collection:
    endpoint: POST /v2/videos/collections/{{ record.video_collection_id }}
    required fields: video_collection_id, name
    risk: external Shutterstock POST /v2/videos/collections/{{ record.video_collection_id }}; approval required
  delete_video_collection:
    endpoint: DELETE /v2/videos/collections/{{ record.video_collection_id }}
    required fields: video_collection_id
    risk: destructive external Shutterstock DELETE /v2/videos/collections/{{ record.video_collection_id }}; approval required
  add_video_collection_items:
    endpoint: POST /v2/videos/collections/{{ record.video_collection_id }}/items
    required fields: video_collection_id, items
    risk: external Shutterstock POST /v2/videos/collections/{{ record.video_collection_id }}/items; approval required
  create_audio_collection:
    endpoint: POST /v2/audio/collections
    required fields: name
    risk: external Shutterstock POST /v2/audio/collections; approval required
  rename_audio_collection:
    endpoint: POST /v2/audio/collections/{{ record.audio_collection_id }}
    required fields: audio_collection_id, name
    risk: external Shutterstock POST /v2/audio/collections/{{ record.audio_collection_id }}; approval required
  delete_audio_collection:
    endpoint: DELETE /v2/audio/collections/{{ record.audio_collection_id }}
    required fields: audio_collection_id
    risk: destructive external Shutterstock DELETE /v2/audio/collections/{{ record.audio_collection_id }}; approval required
  add_audio_collection_items:
    endpoint: POST /v2/audio/collections/{{ record.audio_collection_id }}/items
    required fields: audio_collection_id, items
    risk: external Shutterstock POST /v2/audio/collections/{{ record.audio_collection_id }}/items; approval required
  create_catalog_collection:
    endpoint: POST /v2/catalog/collections
    required fields: name
    risk: external Shutterstock POST /v2/catalog/collections; approval required
  update_catalog_collection:
    endpoint: PATCH /v2/catalog/collections/{{ record.catalog_collection_id }}
    required fields: catalog_collection_id
    risk: external Shutterstock PATCH /v2/catalog/collections/{{ record.catalog_collection_id }}; approval required
  delete_catalog_collection:
    endpoint: DELETE /v2/catalog/collections/{{ record.catalog_collection_id }}
    required fields: catalog_collection_id
    risk: destructive external Shutterstock DELETE /v2/catalog/collections/{{ record.catalog_collection_id }}; approval required
  add_catalog_collection_items:
    endpoint: POST /v2/catalog/collections/{{ record.catalog_collection_id }}/items
    required fields: catalog_collection_id, items
    risk: external Shutterstock POST /v2/catalog/collections/{{ record.catalog_collection_id }}/items; approval required

SECURITY
  read risk: external Shutterstock API read of media, collections, licenses, editorial, catalog, contributor, and subscription metadata
  write risk: external Shutterstock collection/lightbox create, rename, delete, and item-add mutations; licensing/download writes are intentionally excluded
  approval: required before collection write actions; licensing and download endpoints are not exposed as writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Shutterstock's declared streams and reverse-ETL actions.
  Usage: pm shutterstock <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    add audio collection items apply - Plan and execute the add audio collection items reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_audio_collection_items]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/audio/collections/{{ record.audio_collection_id }}/items; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    add catalog collection items apply - Plan and execute the add catalog collection items reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_catalog_collection_items]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/catalog/collections/{{ record.catalog_collection_id }}/items; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    add image collection items apply - Plan and execute the add image collection items reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_image_collection_items]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/images/collections/{{ record.image_collection_id }}/items; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    add video collection items apply - Plan and execute the add video collection items reverse-ETL action [intent=reverse_etl availability=not_implemented write=add_video_collection_items]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/videos/collections/{{ record.video_collection_id }}/items; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    api delete v2 audio collections id items - Documented DELETE /v2/audio/collections/{id}/items (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.delete.v2-audio-collections-id-items]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v2 catalog collections collection-id items - Documented DELETE /v2/catalog/collections/{collection_id}/items (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.delete.v2-catalog-collections-collection-id-items]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v2 images collections id items - Documented DELETE /v2/images/collections/{id}/items (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.delete.v2-images-collections-id-items]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v2 videos collections id items - Documented DELETE /v2/videos/collections/{id}/items (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.delete.v2-videos-collections-id-items]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get v2 audio genres - Documented GET /v2/audio/genres (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-audio-genres]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 audio instruments - Documented GET /v2/audio/instruments (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-audio-instruments]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 audio moods - Documented GET /v2/audio/moods (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-audio-moods]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 cv keywords - Documented GET /v2/cv/keywords (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-cv-keywords]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 editorial categories - Documented GET /v2/editorial/categories (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-editorial-categories]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 editorial images categories - Documented GET /v2/editorial/images/categories (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-editorial-images-categories]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 editorial videos categories - Documented GET /v2/editorial/videos/categories (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-editorial-videos-categories]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 images categories - Documented GET /v2/images/categories (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-images-categories]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 images search suggestions - Documented GET /v2/images/search/suggestions (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-images-search-suggestions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 oauth authorize - Documented GET /v2/oauth/authorize (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-oauth-authorize]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 test - Documented GET /v2/test (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-test]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 test validate - Documented GET /v2/test/validate (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-test-validate]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 user - Documented GET /v2/user (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-user]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 user access-token - Documented GET /v2/user/access_token (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-user-access-token]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 videos categories - Documented GET /v2/videos/categories (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-videos-categories]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 videos search suggestions - Documented GET /v2/videos/search/suggestions (not implemented) [intent=direct_read availability=not_implemented operation=shutterstock.get.v2-videos-search-suggestions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post v2 audio licenses - Documented POST /v2/audio/licenses (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-audio-licenses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 audio licenses id downloads - Documented POST /v2/audio/licenses/{id}/downloads (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-audio-licenses-id-downloads]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 bulk-search images - Documented POST /v2/bulk_search/images (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-bulk-search-images]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 cv images - Documented POST /v2/cv/images (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-cv-images]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 editorial images licenses - Documented POST /v2/editorial/images/licenses (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-editorial-images-licenses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 editorial licenses - Documented POST /v2/editorial/licenses (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-editorial-licenses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 editorial videos licenses - Documented POST /v2/editorial/videos/licenses (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-editorial-videos-licenses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 images licenses - Documented POST /v2/images/licenses (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-images-licenses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 images licenses id downloads - Documented POST /v2/images/licenses/{id}/downloads (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-images-licenses-id-downloads]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 images search suggestions - Documented POST /v2/images/search/suggestions (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-images-search-suggestions]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 oauth access-token - Documented POST /v2/oauth/access_token (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-oauth-access-token]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 sfx licenses - Documented POST /v2/sfx/licenses (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-sfx-licenses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 sfx licenses id downloads - Documented POST /v2/sfx/licenses/{id}/downloads (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-sfx-licenses-id-downloads]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 videos licenses - Documented POST /v2/videos/licenses (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-videos-licenses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 videos licenses id downloads - Documented POST /v2/videos/licenses/{id}/downloads (not implemented) [intent=direct_write availability=not_implemented operation=shutterstock.post.v2-videos-licenses-id-downloads]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    audio list - Run the audio ETL stream [intent=etl availability=implemented stream=audio]
    create audio collection apply - Plan and execute the create audio collection reverse-ETL action [intent=reverse_etl availability=implemented write=create_audio_collection]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/audio/collections; approval required; flags: --name (required)
    create catalog collection apply - Plan and execute the create catalog collection reverse-ETL action [intent=reverse_etl availability=implemented write=create_catalog_collection]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/catalog/collections; approval required; flags: --name (required)
    create image collection apply - Plan and execute the create image collection reverse-ETL action [intent=reverse_etl availability=implemented write=create_image_collection]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/images/collections; approval required; flags: --name (required)
    create video collection apply - Plan and execute the create video collection reverse-ETL action [intent=reverse_etl availability=implemented write=create_video_collection]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/videos/collections; approval required; flags: --name (required)
    delete audio collection apply - Plan and execute the delete audio collection reverse-ETL action [intent=reverse_etl availability=implemented write=delete_audio_collection]; approval: requires plan, preview, approval, and execute; risk: destructive external Shutterstock DELETE /v2/audio/collections/{{ record.audio_collection_id }}; approval required; flags: --audio_collection_id (required)
    delete catalog collection apply - Plan and execute the delete catalog collection reverse-ETL action [intent=reverse_etl availability=implemented write=delete_catalog_collection]; approval: requires plan, preview, approval, and execute; risk: destructive external Shutterstock DELETE /v2/catalog/collections/{{ record.catalog_collection_id }}; approval required; flags: --catalog_collection_id (required)
    delete image collection apply - Plan and execute the delete image collection reverse-ETL action [intent=reverse_etl availability=implemented write=delete_image_collection]; approval: requires plan, preview, approval, and execute; risk: destructive external Shutterstock DELETE /v2/images/collections/{{ record.image_collection_id }}; approval required; flags: --image_collection_id (required)
    delete video collection apply - Plan and execute the delete video collection reverse-ETL action [intent=reverse_etl availability=implemented write=delete_video_collection]; approval: requires plan, preview, approval, and execute; risk: destructive external Shutterstock DELETE /v2/videos/collections/{{ record.video_collection_id }}; approval required; flags: --video_collection_id (required)
    find similar videos list - Run the find similar videos ETL stream [intent=etl availability=implemented stream=find_similar_videos]
    get collections list - Run the get collections ETL stream [intent=etl availability=implemented stream=get_collections]
    get contributor collection items list - Run the get contributor collection items ETL stream [intent=etl availability=implemented stream=get_contributor_collection_items]
    get contributor collection list - Run the get contributor collection ETL stream [intent=etl availability=implemented stream=get_contributor_collection]
    get contributor collections list - Run the get contributor collections ETL stream [intent=etl availability=implemented stream=get_contributor_collections]
    get contributor list - Run the get contributor ETL stream [intent=etl availability=implemented stream=get_contributor]
    get contributors list - Run the get contributors ETL stream [intent=etl availability=implemented stream=get_contributors]
    get editorial image license list list - Run the get editorial image license list ETL stream [intent=etl availability=implemented stream=get_editorial_image_license_list]
    get editorial image list - Run the get editorial image ETL stream [intent=etl availability=implemented stream=get_editorial_image]
    get editorial image list list - Run the get editorial image list ETL stream [intent=etl availability=implemented stream=get_editorial_image_list]
    get editorial image livefeed items list - Run the get editorial image livefeed items ETL stream [intent=etl availability=implemented stream=get_editorial_image_livefeed_items]
    get editorial image livefeed list - Run the get editorial image livefeed ETL stream [intent=etl availability=implemented stream=get_editorial_image_livefeed]
    get editorial image livefeeds list - Run the get editorial image livefeeds ETL stream [intent=etl availability=implemented stream=get_editorial_image_livefeeds]
    get editorial list - Run the get editorial ETL stream [intent=etl availability=implemented stream=get_editorial]
    get editorial livefeed items list - Run the get editorial livefeed items ETL stream [intent=etl availability=implemented stream=get_editorial_livefeed_items]
    get editorial livefeed list - Run the get editorial livefeed ETL stream [intent=etl availability=implemented stream=get_editorial_livefeed]
    get editorial livefeeds list - Run the get editorial livefeeds ETL stream [intent=etl availability=implemented stream=get_editorial_livefeeds]
    get editorial video license list list - Run the get editorial video license list ETL stream [intent=etl availability=implemented stream=get_editorial_video_license_list]
    get editorial video list - Run the get editorial video ETL stream [intent=etl availability=implemented stream=get_editorial_video]
    get editorial video list list - Run the get editorial video list ETL stream [intent=etl availability=implemented stream=get_editorial_video_list]
    get image collection items list - Run the get image collection items ETL stream [intent=etl availability=implemented stream=get_image_collection_items]
    get image collection list - Run the get image collection ETL stream [intent=etl availability=implemented stream=get_image_collection]
    get image collection list list - Run the get image collection list ETL stream [intent=etl availability=implemented stream=get_image_collection_list]
    get image license list list - Run the get image license list ETL stream [intent=etl availability=implemented stream=get_image_license_list]
    get image list - Run the get image ETL stream [intent=etl availability=implemented stream=get_image]
    get image list list - Run the get image list ETL stream [intent=etl availability=implemented stream=get_image_list]
    get image recommendations list - Run the get image recommendations ETL stream [intent=etl availability=implemented stream=get_image_recommendations]
    get sfx license list list - Run the get sfx license list ETL stream [intent=etl availability=implemented stream=get_sfx_license_list]
    get sfx list - Run the get sfx ETL stream [intent=etl availability=implemented stream=get_sfx]
    get sfx list list - Run the get sfx list ETL stream [intent=etl availability=implemented stream=get_sfx_list]
    get similar cv images list - Run the get similar cv images ETL stream [intent=etl availability=implemented stream=get_similar_cv_images]
    get similar cv videos list - Run the get similar cv videos ETL stream [intent=etl availability=implemented stream=get_similar_cv_videos]
    get track collection items list - Run the get track collection items ETL stream [intent=etl availability=implemented stream=get_track_collection_items]
    get track collection list - Run the get track collection ETL stream [intent=etl availability=implemented stream=get_track_collection]
    get track collection list list - Run the get track collection list ETL stream [intent=etl availability=implemented stream=get_track_collection_list]
    get track license list list - Run the get track license list ETL stream [intent=etl availability=implemented stream=get_track_license_list]
    get track list - Run the get track ETL stream [intent=etl availability=implemented stream=get_track]
    get track list list - Run the get track list ETL stream [intent=etl availability=implemented stream=get_track_list]
    get updated editorial images list - Run the get updated editorial images ETL stream [intent=etl availability=implemented stream=get_updated_editorial_images]
    get updated editorial list - Run the get updated editorial ETL stream [intent=etl availability=implemented stream=get_updated_editorial]
    get updated images list - Run the get updated images ETL stream [intent=etl availability=implemented stream=get_updated_images]
    get updated videos list - Run the get updated videos ETL stream [intent=etl availability=implemented stream=get_updated_videos]
    get user subscriptions list - Run the get user subscriptions ETL stream [intent=etl availability=implemented stream=get_user_subscriptions]
    get video collection items list - Run the get video collection items ETL stream [intent=etl availability=implemented stream=get_video_collection_items]
    get video collection list - Run the get video collection ETL stream [intent=etl availability=implemented stream=get_video_collection]
    get video collection list list - Run the get video collection list ETL stream [intent=etl availability=implemented stream=get_video_collection_list]
    get video license list list - Run the get video license list ETL stream [intent=etl availability=implemented stream=get_video_license_list]
    get video list - Run the get video ETL stream [intent=etl availability=implemented stream=get_video]
    get video list list - Run the get video list ETL stream [intent=etl availability=implemented stream=get_video_list]
    images list - Run the images ETL stream [intent=etl availability=implemented stream=images]
    list similar images list - Run the list similar images ETL stream [intent=etl availability=implemented stream=list_similar_images]
    rename audio collection apply - Plan and execute the rename audio collection reverse-ETL action [intent=reverse_etl availability=implemented write=rename_audio_collection]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/audio/collections/{{ record.audio_collection_id }}; approval required; flags: --audio_collection_id (required), --name (required)
    rename image collection apply - Plan and execute the rename image collection reverse-ETL action [intent=reverse_etl availability=implemented write=rename_image_collection]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/images/collections/{{ record.image_collection_id }}; approval required; flags: --image_collection_id (required), --name (required)
    rename video collection apply - Plan and execute the rename video collection reverse-ETL action [intent=reverse_etl availability=implemented write=rename_video_collection]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock POST /v2/videos/collections/{{ record.video_collection_id }}; approval required; flags: --name (required), --video_collection_id (required)
    search catalog list - Run the search catalog ETL stream [intent=etl availability=implemented stream=search_catalog]
    search editorial images list - Run the search editorial images ETL stream [intent=etl availability=implemented stream=search_editorial_images]
    search editorial list - Run the search editorial ETL stream [intent=etl availability=implemented stream=search_editorial]
    search editorial videos list - Run the search editorial videos ETL stream [intent=etl availability=implemented stream=search_editorial_videos]
    search sfx list - Run the search sfx ETL stream [intent=etl availability=implemented stream=search_sfx]
    update catalog collection apply - Plan and execute the update catalog collection reverse-ETL action [intent=reverse_etl availability=implemented write=update_catalog_collection]; approval: requires plan, preview, approval, and execute; risk: external Shutterstock PATCH /v2/catalog/collections/{{ record.catalog_collection_id }}; approval required; flags: --catalog_collection_id (required)
    videos list - Run the videos ETL stream [intent=etl availability=implemented stream=videos]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect shutterstock

  # Inspect as structured JSON
  pm connectors inspect shutterstock --json

AGENT WORKFLOW
  - Run pm connectors inspect shutterstock before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
