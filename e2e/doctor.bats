#!/usr/bin/env bats
load "test_helper"

@test "doctor exits 1 with no config" {
    # No create_global_config — config is absent
    run "$STQRY" doctor
    assert_failure
    assert_output_contains "Config"
    assert_output_contains "✗"
    assert_output_contains "Global config exists"
}

@test "doctor always shows all three sections" {
    run "$STQRY" doctor
    assert_failure
    assert_output_contains "Config"
    assert_output_contains "API"
    assert_output_contains "Version"
}

@test "doctor shows expected check names" {
    run "$STQRY" doctor
    assert_failure
    assert_output_contains "Global config exists"
    assert_output_contains "API reachable"
    assert_output_contains "CLI version"
}

@test "doctor --verbose shows duration in each check line" {
    run "$STQRY" doctor --verbose
    assert_failure
    # Verbose mode shows duration as "{name} ({duration})" e.g. "Global config exists (0s)"
    assert_output_contains "Global config exists ("
    assert_output_contains "API reachable ("
}

@test "doctor with valid config shows config checks passing" {
    create_global_config
    run "$STQRY" doctor --site=testsite
    # Config checks pass; API checks fail (no server at localhost:8765) → overall exit 1
    assert_failure
    assert_output_contains "✓"
    assert_output_contains "Global config exists"
    # Confirm it fails for the right reason (API unreachable), not a crash or config error
    assert_output_contains "API reachable"
}

@test "doctor --help shows usage" {
    run "$STQRY" doctor --help
    assert_success
    assert_output_contains "doctor"
    assert_output_contains "--verbose"
}

# Happy-path tests: replay proxy serves cassettes so all API checks pass.
# The binary exits 0 only when every check is pass/warn/info (no fail).

@test "doctor exits 0 when all checks pass" {
    create_global_config
    create_directory_config
    run "$STQRY" doctor --site=testsite
    assert_success
    assert_output_contains "✓"
    assert_output_contains "Global config exists"
    assert_output_contains "API reachable"
    assert_output_contains "Token valid"
}

@test "doctor --verbose exits 0 and shows durations" {
    create_global_config
    create_directory_config
    run "$STQRY" doctor --verbose --site=testsite
    assert_success
    assert_output_contains "Global config exists ("
    assert_output_contains "API reachable ("
    assert_output_contains "Token valid ("
}
