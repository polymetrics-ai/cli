#!/usr/bin/env bash
# Immutable Phase 0 safety policy. No environment, argument, state, or prompt
# can enable autonomous run/resume through this file.

readonly AUTO_LOOP_SAFETY_SCHEMA_VERSION="1.0"
readonly AUTO_LOOP_SAFETY_STATE="closed"
readonly AUTO_LOOP_SAFETY_CODE="AUTO_LOOP_DISABLED_PHASE_0"
readonly AUTO_LOOP_SAFETY_EXIT_CLASS="safety_disabled"

auto_loop_safety_status_json() {
  printf '{"schema_version":"%s","state":"%s","run_enabled":false,"resume_enabled":false,"code":"%s","exit_class":"%s"}\n' \
    "$AUTO_LOOP_SAFETY_SCHEMA_VERSION" "$AUTO_LOOP_SAFETY_STATE" "$AUTO_LOOP_SAFETY_CODE" "$AUTO_LOOP_SAFETY_EXIT_CLASS"
}

auto_loop_safety_entrypoints() {
  printf '%s\n' \
    'scripts/claude-auto-loop.sh' \
    'scripts/pi-auto-loop.sh' \
    'scripts/pi-shepherd-loop.sh'
}

auto_loop_safety_is_tracked() {
  case "${1:-}" in
    scripts/claude-auto-loop.sh|scripts/pi-auto-loop.sh|scripts/pi-shepherd-loop.sh) return 0 ;;
    *) return 1 ;;
  esac
}

auto_loop_safety_guard_driver() {
  local entrypoint="${1:-}"
  local output="${2:-text}"

  if ! auto_loop_safety_is_tracked "$entrypoint"; then
    if [[ "$output" == "json" ]]; then
      printf '{"schema_version":"%s","state":"closed","run_enabled":false,"resume_enabled":false,"code":"AUTO_LOOP_ENTRYPOINT_UNTRACKED","exit_class":"usage_error"}\n' \
        "$AUTO_LOOP_SAFETY_SCHEMA_VERSION" >&2
    else
      printf 'AUTO_LOOP_ENTRYPOINT_UNTRACKED\n' >&2
    fi
    return 64
  fi

  if [[ "$output" == "json" ]]; then
    printf '{"schema_version":"%s","state":"%s","run_enabled":false,"resume_enabled":false,"code":"%s","exit_class":"%s","entrypoint":"%s"}\n' \
      "$AUTO_LOOP_SAFETY_SCHEMA_VERSION" "$AUTO_LOOP_SAFETY_STATE" "$AUTO_LOOP_SAFETY_CODE" \
      "$AUTO_LOOP_SAFETY_EXIT_CLASS" "$entrypoint" >&2
  else
    printf '%s\n' "$AUTO_LOOP_SAFETY_CODE" >&2
  fi
  return 78
}

auto_loop_safety_usage() {
  printf '%s\n' \
    'usage:' \
    '  scripts/auto-loop-safety.sh status [--json]' \
    '  scripts/auto-loop-safety.sh entrypoints [--json]' \
    '  scripts/auto-loop-safety.sh guard-driver <entrypoint> [--json]'
}

auto_loop_safety_main() {
  local command="${1:-help}"
  case "$command" in
    help|-h|--help)
      auto_loop_safety_usage
      return 0
      ;;
    status)
      case "${2:-}" in
        "") printf 'agent loop safety: closed (run=false resume=false code=%s)\n' "$AUTO_LOOP_SAFETY_CODE" ;;
        --json) auto_loop_safety_status_json ;;
        *) printf 'auto-loop-safety: unknown status argument\n' >&2; return 64 ;;
      esac
      return 0
      ;;
    entrypoints)
      case "${2:-}" in
        "") auto_loop_safety_entrypoints ;;
        --json) printf '{"schema_version":"%s","entrypoints":["scripts/claude-auto-loop.sh","scripts/pi-auto-loop.sh","scripts/pi-shepherd-loop.sh"]}\n' "$AUTO_LOOP_SAFETY_SCHEMA_VERSION" ;;
        *) printf 'auto-loop-safety: unknown entrypoints argument\n' >&2; return 64 ;;
      esac
      return 0
      ;;
    guard-driver)
      local output="text"
      if [[ "${3:-}" == "--json" ]]; then
        output="json"
      elif [[ -n "${3:-}" ]]; then
        printf 'auto-loop-safety: unknown guard-driver argument\n' >&2
        return 64
      fi
      auto_loop_safety_guard_driver "${2:-}" "$output"
      return $?
      ;;
    *)
      printf 'auto-loop-safety: unknown command\n' >&2
      return 64
      ;;
  esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  auto_loop_safety_main "$@"
  exit $?
fi
