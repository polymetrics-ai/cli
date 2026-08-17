---
name: pm-pokeapi
description: PokeAPI connector knowledge and safe action guide.
---

# pm-pokeapi

## Purpose

Reads the documented public PokeAPI v2 resource catalog, including list and detail endpoints.

## Icon

- asset: icons/pokeapi.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://pokeapi.co/docs/v2

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- No secret authentication is required for this connector.

## Configuration

- ability_id
- base_url
- berry_firmness_id
- berry_flavor_id
- berry_id
- characteristic_id
- contest_effect_id
- contest_type_id
- egg_group_id
- encounter_condition_id
- encounter_condition_value_id
- encounter_method_id
- evolution_chain_id
- evolution_trigger_id
- gender_id
- generation_id
- growth_rate_id
- item_attribute_id
- item_category_id
- item_fling_effect_id
- item_id
- item_pocket_id
- language_id
- location_area_id
- location_id
- machine_id
- mode
- move_ailment_id
- move_battle_style_id
- move_category_id
- move_damage_class_id
- move_id
- move_learn_method_id
- move_target_id
- nature_id
- pal_park_area_id
- pokeathlon_stat_id
- pokedex_id
- pokemon_color_id
- pokemon_form_id
- pokemon_habitat_id
- pokemon_id
- pokemon_shape_id
- pokemon_species_id
- region_id
- stat_id
- super_contest_effect_id
- type_id
- version_group_id
- version_id

## ETL Streams

- pokemon:
  - primary key: name
  - fields: id(), name(), url()
- types:
  - primary key: name
  - fields: id(), name(), url()
- abilities:
  - primary key: name
  - fields: id(), name(), url()
- moves:
  - primary key: name
  - fields: id(), name(), url()
- berries:
  - primary key: name
  - fields: id(), name(), url()
- berry_firmnesses:
  - primary key: name
  - fields: id(), name(), url()
- berry_flavors:
  - primary key: name
  - fields: id(), name(), url()
- contest_types:
  - primary key: name
  - fields: id(), name(), url()
- contest_effects:
  - primary key: id
  - fields: id(), url()
- super_contest_effects:
  - primary key: id
  - fields: id(), url()
- encounter_methods:
  - primary key: name
  - fields: id(), name(), url()
- encounter_conditions:
  - primary key: name
  - fields: id(), name(), url()
- encounter_condition_values:
  - primary key: name
  - fields: id(), name(), url()
- evolution_chains:
  - primary key: id
  - fields: id(), url()
- evolution_triggers:
  - primary key: name
  - fields: id(), name(), url()
- generations:
  - primary key: name
  - fields: id(), name(), url()
- pokedexes:
  - primary key: name
  - fields: id(), name(), url()
- versions:
  - primary key: name
  - fields: id(), name(), url()
- version_groups:
  - primary key: name
  - fields: id(), name(), url()
- items:
  - primary key: name
  - fields: id(), name(), url()
- item_attributes:
  - primary key: name
  - fields: id(), name(), url()
- item_categories:
  - primary key: name
  - fields: id(), name(), url()
- item_fling_effects:
  - primary key: name
  - fields: id(), name(), url()
- item_pockets:
  - primary key: name
  - fields: id(), name(), url()
- locations:
  - primary key: name
  - fields: id(), name(), url()
- location_areas:
  - primary key: name
  - fields: id(), name(), url()
- pal_park_areas:
  - primary key: name
  - fields: id(), name(), url()
- regions:
  - primary key: name
  - fields: id(), name(), url()
- machines:
  - primary key: id
  - fields: id(), url()
- move_ailments:
  - primary key: name
  - fields: id(), name(), url()
- move_battle_styles:
  - primary key: name
  - fields: id(), name(), url()
- move_categories:
  - primary key: name
  - fields: id(), name(), url()
- move_damage_classes:
  - primary key: name
  - fields: id(), name(), url()
- move_learn_methods:
  - primary key: name
  - fields: id(), name(), url()
- move_targets:
  - primary key: name
  - fields: id(), name(), url()
- characteristics:
  - primary key: id
  - fields: id(), url()
- egg_groups:
  - primary key: name
  - fields: id(), name(), url()
- genders:
  - primary key: name
  - fields: id(), name(), url()
- growth_rates:
  - primary key: name
  - fields: id(), name(), url()
- natures:
  - primary key: name
  - fields: id(), name(), url()
- pokeathlon_stats:
  - primary key: name
  - fields: id(), name(), url()
- pokemon_colors:
  - primary key: name
  - fields: id(), name(), url()
- pokemon_forms:
  - primary key: name
  - fields: id(), name(), url()
- pokemon_habitats:
  - primary key: name
  - fields: id(), name(), url()
- pokemon_shapes:
  - primary key: name
  - fields: id(), name(), url()
- pokemon_species:
  - primary key: name
  - fields: id(), name(), url()
- stats:
  - primary key: name
  - fields: id(), name(), url()
- languages:
  - primary key: name
  - fields: id(), name(), url()
- pokemon_detail:
  - primary key: id
  - fields: id(), name()
- types_detail:
  - primary key: id
  - fields: id(), name()
- abilities_detail:
  - primary key: id
  - fields: id(), name()
- moves_detail:
  - primary key: id
  - fields: id(), name()
- berries_detail:
  - primary key: id
  - fields: id(), name()
- berry_firmnesses_detail:
  - primary key: id
  - fields: id(), name()
- berry_flavors_detail:
  - primary key: id
  - fields: id(), name()
- contest_types_detail:
  - primary key: id
  - fields: id(), name()
- contest_effects_detail:
  - primary key: id
  - fields: id()
- super_contest_effects_detail:
  - primary key: id
  - fields: id()
- encounter_methods_detail:
  - primary key: id
  - fields: id(), name()
- encounter_conditions_detail:
  - primary key: id
  - fields: id(), name()
- encounter_condition_values_detail:
  - primary key: id
  - fields: id(), name()
- evolution_chains_detail:
  - primary key: id
  - fields: id()
- evolution_triggers_detail:
  - primary key: id
  - fields: id(), name()
- generations_detail:
  - primary key: id
  - fields: id(), name()
- pokedexes_detail:
  - primary key: id
  - fields: id(), name()
- versions_detail:
  - primary key: id
  - fields: id(), name()
- version_groups_detail:
  - primary key: id
  - fields: id(), name()
- items_detail:
  - primary key: id
  - fields: id(), name()
- item_attributes_detail:
  - primary key: id
  - fields: id(), name()
- item_categories_detail:
  - primary key: id
  - fields: id(), name()
- item_fling_effects_detail:
  - primary key: id
  - fields: id(), name()
- item_pockets_detail:
  - primary key: id
  - fields: id(), name()
- locations_detail:
  - primary key: id
  - fields: id(), name()
- location_areas_detail:
  - primary key: id
  - fields: id(), name()
- pal_park_areas_detail:
  - primary key: id
  - fields: id(), name()
- regions_detail:
  - primary key: id
  - fields: id(), name()
- machines_detail:
  - primary key: id
  - fields: id()
- move_ailments_detail:
  - primary key: id
  - fields: id(), name()
- move_battle_styles_detail:
  - primary key: id
  - fields: id(), name()
- move_categories_detail:
  - primary key: id
  - fields: id(), name()
- move_damage_classes_detail:
  - primary key: id
  - fields: id(), name()
- move_learn_methods_detail:
  - primary key: id
  - fields: id(), name()
- move_targets_detail:
  - primary key: id
  - fields: id(), name()
- characteristics_detail:
  - primary key: id
  - fields: id()
- egg_groups_detail:
  - primary key: id
  - fields: id(), name()
- genders_detail:
  - primary key: id
  - fields: id(), name()
- growth_rates_detail:
  - primary key: id
  - fields: id(), name()
- natures_detail:
  - primary key: id
  - fields: id(), name()
- pokeathlon_stats_detail:
  - primary key: id
  - fields: id(), name()
- pokemon_colors_detail:
  - primary key: id
  - fields: id(), name()
- pokemon_forms_detail:
  - primary key: id
  - fields: id(), name()
- pokemon_habitats_detail:
  - primary key: id
  - fields: id(), name()
- pokemon_shapes_detail:
  - primary key: id
  - fields: id(), name()
- pokemon_species_detail:
  - primary key: id
  - fields: id(), name()
- stats_detail:
  - primary key: id
  - fields: id(), name()
- languages_detail:
  - primary key: id
  - fields: id(), name()
- pokemon_location_areas:
  - primary key: id
  - fields: id(), location_area(), version_details()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external PokeAPI read of public Pokemon reference data
- approval: none; read-only public reference API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect pokeapi
```

### Inspect as structured JSON

```bash
pm connectors inspect pokeapi --json
```

## Agent Rules

- Run pm connectors inspect pokeapi before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
