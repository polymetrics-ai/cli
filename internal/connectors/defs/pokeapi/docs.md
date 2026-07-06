# Overview

Reads the documented public PokeAPI v2 resource catalog, including list and detail endpoints.

Readable streams: `pokemon`, `types`, `abilities`, `moves`, `berries`, `berry_firmnesses`,
`berry_flavors`, `contest_types`, `contest_effects`, `super_contest_effects`, `encounter_methods`,
`encounter_conditions`, `encounter_condition_values`, `evolution_chains`, `evolution_triggers`,
`generations`, `pokedexes`, `versions`, `version_groups`, `items`, `item_attributes`,
`item_categories`, `item_fling_effects`, `item_pockets`, `locations`, `location_areas`,
`pal_park_areas`, `regions`, `machines`, `move_ailments`, `move_battle_styles`, `move_categories`,
`move_damage_classes`, `move_learn_methods`, `move_targets`, `characteristics`, `egg_groups`,
`genders`, `growth_rates`, `natures`, `pokeathlon_stats`, `pokemon_colors`, `pokemon_forms`,
`pokemon_habitats`, `pokemon_shapes`, `pokemon_species`, `stats`, `languages`, `pokemon_detail`,
`types_detail`, `abilities_detail`, `moves_detail`, `berries_detail`, `berry_firmnesses_detail`,
`berry_flavors_detail`, `contest_types_detail`, `contest_effects_detail`,
`super_contest_effects_detail`, `encounter_methods_detail`, `encounter_conditions_detail`,
`encounter_condition_values_detail`, `evolution_chains_detail`, `evolution_triggers_detail`,
`generations_detail`, `pokedexes_detail`, `versions_detail`, `version_groups_detail`,
`items_detail`, `item_attributes_detail`, `item_categories_detail`, `item_fling_effects_detail`,
`item_pockets_detail`, `locations_detail`, `location_areas_detail`, `pal_park_areas_detail`,
`regions_detail`, `machines_detail`, `move_ailments_detail`, `move_battle_styles_detail`,
`move_categories_detail`, `move_damage_classes_detail`, `move_learn_methods_detail`,
`move_targets_detail`, `characteristics_detail`, `egg_groups_detail`, `genders_detail`,
`growth_rates_detail`, `natures_detail`, `pokeathlon_stats_detail`, `pokemon_colors_detail`,
`pokemon_forms_detail`, `pokemon_habitats_detail`, `pokemon_shapes_detail`,
`pokemon_species_detail`, `stats_detail`, `languages_detail`, `pokemon_location_areas`.

This connector is read-only; no write actions are declared.

Service API documentation: https://pokeapi.co/docs/v2.

## Auth setup

Connection fields:

- `ability_id` (optional, string); ID or name used by the abilities_detail stream.
- `base_url` (optional, string); default `https://pokeapi.co/api/v2`; format `uri`; PokeAPI base URL
  override for tests or proxies.
- `berry_firmness_id` (optional, string); ID or name used by the berry_firmnesses_detail stream.
- `berry_flavor_id` (optional, string); ID or name used by the berry_flavors_detail stream.
- `berry_id` (optional, string); ID or name used by the berries_detail stream.
- `characteristic_id` (optional, string); ID or name used by the characteristics_detail stream.
- `contest_effect_id` (optional, string); ID or name used by the contest_effects_detail stream.
- `contest_type_id` (optional, string); ID or name used by the contest_types_detail stream.
- `egg_group_id` (optional, string); ID or name used by the egg_groups_detail stream.
- `encounter_condition_id` (optional, string); ID or name used by the encounter_conditions_detail
  stream.
- `encounter_condition_value_id` (optional, string); ID or name used by the
  encounter_condition_values_detail stream.
- `encounter_method_id` (optional, string); ID or name used by the encounter_methods_detail stream.
- `evolution_chain_id` (optional, string); ID or name used by the evolution_chains_detail stream.
- `evolution_trigger_id` (optional, string); ID or name used by the evolution_triggers_detail
  stream.
- `gender_id` (optional, string); ID or name used by the genders_detail stream.
- `generation_id` (optional, string); ID or name used by the generations_detail stream.
- `growth_rate_id` (optional, string); ID or name used by the growth_rates_detail stream.
- `item_attribute_id` (optional, string); ID or name used by the item_attributes_detail stream.
- `item_category_id` (optional, string); ID or name used by the item_categories_detail stream.
- `item_fling_effect_id` (optional, string); ID or name used by the item_fling_effects_detail
  stream.
- `item_id` (optional, string); ID or name used by the items_detail stream.
- `item_pocket_id` (optional, string); ID or name used by the item_pockets_detail stream.
- `language_id` (optional, string); ID or name used by the languages_detail stream.
- `location_area_id` (optional, string); ID or name used by the location_areas_detail stream.
- `location_id` (optional, string); ID or name used by the locations_detail stream.
- `machine_id` (optional, string); ID or name used by the machines_detail stream.
- `mode` (optional, string).
- `move_ailment_id` (optional, string); ID or name used by the move_ailments_detail stream.
- `move_battle_style_id` (optional, string); ID or name used by the move_battle_styles_detail
  stream.
- `move_category_id` (optional, string); ID or name used by the move_categories_detail stream.
- `move_damage_class_id` (optional, string); ID or name used by the move_damage_classes_detail
  stream.
- `move_id` (optional, string); ID or name used by the moves_detail stream.
- `move_learn_method_id` (optional, string); ID or name used by the move_learn_methods_detail
  stream.
- `move_target_id` (optional, string); ID or name used by the move_targets_detail stream.
- `nature_id` (optional, string); ID or name used by the natures_detail stream.
- `pal_park_area_id` (optional, string); ID or name used by the pal_park_areas_detail stream.
- `pokeathlon_stat_id` (optional, string); ID or name used by the pokeathlon_stats_detail stream.
- `pokedex_id` (optional, string); ID or name used by the pokedexes_detail stream.
- `pokemon_color_id` (optional, string); ID or name used by the pokemon_colors_detail stream.
- `pokemon_form_id` (optional, string); ID or name used by the pokemon_forms_detail stream.
- `pokemon_habitat_id` (optional, string); ID or name used by the pokemon_habitats_detail stream.
- `pokemon_id` (optional, string); ID or name used by the pokemon_detail stream.
- `pokemon_shape_id` (optional, string); ID or name used by the pokemon_shapes_detail stream.
- `pokemon_species_id` (optional, string); ID or name used by the pokemon_species_detail stream.
- `region_id` (optional, string); ID or name used by the regions_detail stream.
- `stat_id` (optional, string); ID or name used by the stats_detail stream.
- `super_contest_effect_id` (optional, string); ID or name used by the super_contest_effects_detail
  stream.
- `type_id` (optional, string); ID or name used by the types_detail stream.
- `version_group_id` (optional, string); ID or name used by the version_groups_detail stream.
- `version_id` (optional, string); ID or name used by the versions_detail stream.

Default configuration values: `base_url=https://pokeapi.co/api/v2`.

Authentication behavior:

- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/pokemon` with query `limit`=`1`; `offset`=`0`.

## Streams notes

Default pagination: offset/limit pagination; offset parameter `offset`; limit parameter `limit`;
page size 100; maximum 3 page(s).

Pagination by stream: none: `pokemon_detail`, `types_detail`, `abilities_detail`, `moves_detail`,
`berries_detail`, `berry_firmnesses_detail`, `berry_flavors_detail`, `contest_types_detail`,
`contest_effects_detail`, `super_contest_effects_detail`, `encounter_methods_detail`,
`encounter_conditions_detail`, `encounter_condition_values_detail`, `evolution_chains_detail`,
`evolution_triggers_detail`, `generations_detail`, `pokedexes_detail`, `versions_detail`,
`version_groups_detail`, `items_detail`, `item_attributes_detail`, `item_categories_detail`,
`item_fling_effects_detail`, `item_pockets_detail`, `locations_detail`, `location_areas_detail`,
`pal_park_areas_detail`, `regions_detail`, `machines_detail`, `move_ailments_detail`,
`move_battle_styles_detail`, `move_categories_detail`, `move_damage_classes_detail`,
`move_learn_methods_detail`, `move_targets_detail`, `characteristics_detail`, `egg_groups_detail`,
`genders_detail`, `growth_rates_detail`, `natures_detail`, `pokeathlon_stats_detail`,
`pokemon_colors_detail`, `pokemon_forms_detail`, `pokemon_habitats_detail`, `pokemon_shapes_detail`,
`pokemon_species_detail`, `stats_detail`, `languages_detail`, `pokemon_location_areas`;
offset_limit: `pokemon`, `types`, `abilities`, `moves`, `berries`, `berry_firmnesses`,
`berry_flavors`, `contest_types`, `contest_effects`, `super_contest_effects`, `encounter_methods`,
`encounter_conditions`, `encounter_condition_values`, `evolution_chains`, `evolution_triggers`,
`generations`, `pokedexes`, `versions`, `version_groups`, `items`, `item_attributes`,
`item_categories`, `item_fling_effects`, `item_pockets`, `locations`, `location_areas`,
`pal_park_areas`, `regions`, `machines`, `move_ailments`, `move_battle_styles`, `move_categories`,
`move_damage_classes`, `move_learn_methods`, `move_targets`, `characteristics`, `egg_groups`,
`genders`, `growth_rates`, `natures`, `pokeathlon_stats`, `pokemon_colors`, `pokemon_forms`,
`pokemon_habitats`, `pokemon_shapes`, `pokemon_species`, `stats`, `languages`.

- `pokemon`: GET `/pokemon` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `types`: GET `/type` - records path `results`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `abilities`: GET `/ability` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `moves`: GET `/move` - records path `results`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `berries`: GET `/berry` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `berry_firmnesses`: GET `/berry-firmness` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `berry_flavors`: GET `/berry-flavor` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `contest_types`: GET `/contest-type` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `contest_effects`: GET `/contest-effect` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `super_contest_effects`: GET `/super-contest-effect` - records path `results`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s);
  computed output fields `id`.
- `encounter_methods`: GET `/encounter-method` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `encounter_conditions`: GET `/encounter-condition` - records path `results`; offset/limit
  pagination; offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s);
  computed output fields `id`.
- `encounter_condition_values`: GET `/encounter-condition-value` - records path `results`;
  offset/limit pagination; offset parameter `offset`; limit parameter `limit`; page size 100;
  maximum 3 page(s); computed output fields `id`.
- `evolution_chains`: GET `/evolution-chain` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `evolution_triggers`: GET `/evolution-trigger` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `generations`: GET `/generation` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `pokedexes`: GET `/pokedex` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `versions`: GET `/version` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `version_groups`: GET `/version-group` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `items`: GET `/item` - records path `results`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `item_attributes`: GET `/item-attribute` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `item_categories`: GET `/item-category` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `item_fling_effects`: GET `/item-fling-effect` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `item_pockets`: GET `/item-pocket` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `locations`: GET `/location` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `location_areas`: GET `/location-area` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `pal_park_areas`: GET `/pal-park-area` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `regions`: GET `/region` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `machines`: GET `/machine` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `move_ailments`: GET `/move-ailment` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `move_battle_styles`: GET `/move-battle-style` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `move_categories`: GET `/move-category` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `move_damage_classes`: GET `/move-damage-class` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `move_learn_methods`: GET `/move-learn-method` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `move_targets`: GET `/move-target` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `characteristics`: GET `/characteristic` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `egg_groups`: GET `/egg-group` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `genders`: GET `/gender` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `growth_rates`: GET `/growth-rate` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `natures`: GET `/nature` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `pokeathlon_stats`: GET `/pokeathlon-stat` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `pokemon_colors`: GET `/pokemon-color` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `pokemon_forms`: GET `/pokemon-form` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `pokemon_habitats`: GET `/pokemon-habitat` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `pokemon_shapes`: GET `/pokemon-shape` - records path `results`; offset/limit pagination; offset
  parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output
  fields `id`.
- `pokemon_species`: GET `/pokemon-species` - records path `results`; offset/limit pagination;
  offset parameter `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed
  output fields `id`.
- `stats`: GET `/stat` - records path `results`; offset/limit pagination; offset parameter `offset`;
  limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `languages`: GET `/language` - records path `results`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 100; maximum 3 page(s); computed output fields `id`.
- `pokemon_detail`: GET `/pokemon/{{ config.pokemon_id }}` - records at response root; emits
  passthrough records.
- `types_detail`: GET `/type/{{ config.type_id }}` - records at response root; emits passthrough
  records.
- `abilities_detail`: GET `/ability/{{ config.ability_id }}` - records at response root; emits
  passthrough records.
- `moves_detail`: GET `/move/{{ config.move_id }}` - records at response root; emits passthrough
  records.
- `berries_detail`: GET `/berry/{{ config.berry_id }}` - records at response root; emits passthrough
  records.
- `berry_firmnesses_detail`: GET `/berry-firmness/{{ config.berry_firmness_id }}` - records at
  response root; emits passthrough records.
- `berry_flavors_detail`: GET `/berry-flavor/{{ config.berry_flavor_id }}` - records at response
  root; emits passthrough records.
- `contest_types_detail`: GET `/contest-type/{{ config.contest_type_id }}` - records at response
  root; emits passthrough records.
- `contest_effects_detail`: GET `/contest-effect/{{ config.contest_effect_id }}` - records at
  response root; emits passthrough records.
- `super_contest_effects_detail`: GET `/super-contest-effect/{{ config.super_contest_effect_id }}` -
  records at response root; emits passthrough records.
- `encounter_methods_detail`: GET `/encounter-method/{{ config.encounter_method_id }}` - records at
  response root; emits passthrough records.
- `encounter_conditions_detail`: GET `/encounter-condition/{{ config.encounter_condition_id }}` -
  records at response root; emits passthrough records.
- `encounter_condition_values_detail`: GET `/encounter-condition-value/{{
  config.encounter_condition_value_id }}` - records at response root; emits passthrough records.
- `evolution_chains_detail`: GET `/evolution-chain/{{ config.evolution_chain_id }}` - records at
  response root; emits passthrough records.
- `evolution_triggers_detail`: GET `/evolution-trigger/{{ config.evolution_trigger_id }}` - records
  at response root; emits passthrough records.
- `generations_detail`: GET `/generation/{{ config.generation_id }}` - records at response root;
  emits passthrough records.
- `pokedexes_detail`: GET `/pokedex/{{ config.pokedex_id }}` - records at response root; emits
  passthrough records.
- `versions_detail`: GET `/version/{{ config.version_id }}` - records at response root; emits
  passthrough records.
- `version_groups_detail`: GET `/version-group/{{ config.version_group_id }}` - records at response
  root; emits passthrough records.
- `items_detail`: GET `/item/{{ config.item_id }}` - records at response root; emits passthrough
  records.
- `item_attributes_detail`: GET `/item-attribute/{{ config.item_attribute_id }}` - records at
  response root; emits passthrough records.
- `item_categories_detail`: GET `/item-category/{{ config.item_category_id }}` - records at response
  root; emits passthrough records.
- `item_fling_effects_detail`: GET `/item-fling-effect/{{ config.item_fling_effect_id }}` - records
  at response root; emits passthrough records.
- `item_pockets_detail`: GET `/item-pocket/{{ config.item_pocket_id }}` - records at response root;
  emits passthrough records.
- `locations_detail`: GET `/location/{{ config.location_id }}` - records at response root; emits
  passthrough records.
- `location_areas_detail`: GET `/location-area/{{ config.location_area_id }}` - records at response
  root; emits passthrough records.
- `pal_park_areas_detail`: GET `/pal-park-area/{{ config.pal_park_area_id }}` - records at response
  root; emits passthrough records.
- `regions_detail`: GET `/region/{{ config.region_id }}` - records at response root; emits
  passthrough records.
- `machines_detail`: GET `/machine/{{ config.machine_id }}` - records at response root; emits
  passthrough records.
- `move_ailments_detail`: GET `/move-ailment/{{ config.move_ailment_id }}` - records at response
  root; emits passthrough records.
- `move_battle_styles_detail`: GET `/move-battle-style/{{ config.move_battle_style_id }}` - records
  at response root; emits passthrough records.
- `move_categories_detail`: GET `/move-category/{{ config.move_category_id }}` - records at response
  root; emits passthrough records.
- `move_damage_classes_detail`: GET `/move-damage-class/{{ config.move_damage_class_id }}` - records
  at response root; emits passthrough records.
- `move_learn_methods_detail`: GET `/move-learn-method/{{ config.move_learn_method_id }}` - records
  at response root; emits passthrough records.
- `move_targets_detail`: GET `/move-target/{{ config.move_target_id }}` - records at response root;
  emits passthrough records.
- `characteristics_detail`: GET `/characteristic/{{ config.characteristic_id }}` - records at
  response root; emits passthrough records.
- `egg_groups_detail`: GET `/egg-group/{{ config.egg_group_id }}` - records at response root; emits
  passthrough records.
- `genders_detail`: GET `/gender/{{ config.gender_id }}` - records at response root; emits
  passthrough records.
- `growth_rates_detail`: GET `/growth-rate/{{ config.growth_rate_id }}` - records at response root;
  emits passthrough records.
- `natures_detail`: GET `/nature/{{ config.nature_id }}` - records at response root; emits
  passthrough records.
- `pokeathlon_stats_detail`: GET `/pokeathlon-stat/{{ config.pokeathlon_stat_id }}` - records at
  response root; emits passthrough records.
- `pokemon_colors_detail`: GET `/pokemon-color/{{ config.pokemon_color_id }}` - records at response
  root; emits passthrough records.
- `pokemon_forms_detail`: GET `/pokemon-form/{{ config.pokemon_form_id }}` - records at response
  root; emits passthrough records.
- `pokemon_habitats_detail`: GET `/pokemon-habitat/{{ config.pokemon_habitat_id }}` - records at
  response root; emits passthrough records.
- `pokemon_shapes_detail`: GET `/pokemon-shape/{{ config.pokemon_shape_id }}` - records at response
  root; emits passthrough records.
- `pokemon_species_detail`: GET `/pokemon-species/{{ config.pokemon_species_id }}` - records at
  response root; emits passthrough records.
- `stats_detail`: GET `/stat/{{ config.stat_id }}` - records at response root; emits passthrough
  records.
- `languages_detail`: GET `/language/{{ config.language_id }}` - records at response root; emits
  passthrough records.
- `pokemon_location_areas`: GET `/pokemon/{{ config.pokemon_id }}/encounters` - records at response
  root; computed output fields `id`; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external PokeAPI read of public Pokemon reference data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 97 stream-backed endpoint group(s).
