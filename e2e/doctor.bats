#!/usr/bin/env bats
load "test_helper"

# Note: tests that assert exit 0 (all checks passing) require a live API server
# and are not included. The happy-path exit code is covered by e2e/collections.bats
# patterns; here we verify behaviour under the most common support scenario (no/bad config).

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
    # Verbose mode shows duration as "{name} ({duration})" e.g. "Global config exists (0s)"
    assert_output_contains "Global config exists ("
}

@test "doctor with valid config shows config checks passing" {
    create_global_config
    run "$STQRY" doctor --site=testsite
    # Config checks pass; API checks fail (no server at localhost:8765) → overall exit 1
    assert_failure
    assert_output_contains "✓"
    assert_output_contains "Global config exists"
}

@test "doctor --help shows usage" {
    run "$STQRY" doctor --help
    assert_success
    assert_output_contains "doctor"
    assert_output_contains "verbose"
}
