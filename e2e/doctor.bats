#!/usr/bin/env bats
load "test_helper"

@test "doctor exits 1 with no config" {
    # No create_global_config — config is absent
    run "$STQRY" doctor
    assert_failure
    assert_output_contains "Config"
    assert_output_contains "✗"
}

@test "doctor shows Config section" {
    run "$STQRY" doctor
    assert_output_contains "Config"
    assert_output_contains "Global config exists"
}

@test "doctor shows API section" {
    run "$STQRY" doctor
    assert_output_contains "API"
    assert_output_contains "API reachable"
}

@test "doctor shows Version section" {
    run "$STQRY" doctor
    assert_output_contains "Version"
    assert_output_contains "CLI version"
}

@test "doctor --verbose shows detail lines" {
    run "$STQRY" doctor --verbose
    # Verbose mode shows duration in parentheses, e.g. "(0s)" or "(12ms)"
    assert_output_contains "s)"
}

@test "doctor with valid config shows config passing" {
    create_global_config
    run "$STQRY" doctor --site=testsite
    # Config checks should pass even if API is unreachable
    assert_output_contains "✓"
    assert_output_contains "Global config exists"
}

@test "doctor --help shows usage" {
    run "$STQRY" doctor --help
    assert_success
    assert_output_contains "doctor"
    assert_output_contains "verbose"
}
