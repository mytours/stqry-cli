#!/usr/bin/env bats
load "test_helper"

@test "config add-site adds a site to global config" {
    run "$STQRY" config add-site --name=mysite --token=abc123 --region=us
    assert_success
    # Verify config was written
    run cat "$TEST_HOME/.config/stqry/config.yaml"
    assert_success
    assert_output_contains "mysite"
    assert_output_contains "abc123"
}

@test "config add-site fails without --token" {
    run "$STQRY" config add-site --name=mysite --region=us
    assert_failure
}

@test "config add-site fails without --region or --api-url" {
    run "$STQRY" config add-site --name=mysite --token=abc123
    assert_failure
}

@test "config add-site fails with unknown region" {
    run "$STQRY" config add-site --name=mysite --token=abc123 --region=xx
    assert_failure
    assert_output_contains "unknown region"
}

@test "config list-sites shows configured sites" {
    create_global_config
    run "$STQRY" config list-sites --site=testsite
    assert_success
    assert_output_contains "testsite"
}

@test "config remove-site removes a site" {
    create_global_config
    run "$STQRY" config remove-site testsite
    assert_success
}

@test "config init --name creates stqry.yaml, AGENTS.md, and CLAUDE.md" {
    create_global_config
    run "$STQRY" config init --name=testsite
    assert_success
    [ -f "$TEST_WORK/stqry.yaml" ]
    [ -f "$TEST_WORK/AGENTS.md" ]
    [ -f "$TEST_WORK/CLAUDE.md" ]
    assert_output_contains "Initialised stqry.yaml for site"
    assert_output_contains "wrote AGENTS.md and CLAUDE.md"
}

@test "config init --token --region creates stqry.yaml, AGENTS.md, and CLAUDE.md" {
    run "$STQRY" config init --token=tok123 --region=us
    assert_success
    [ -f "$TEST_WORK/stqry.yaml" ]
    [ -f "$TEST_WORK/AGENTS.md" ]
    [ -f "$TEST_WORK/CLAUDE.md" ]
    assert_output_contains "Initialised stqry.yaml with inline credentials"
    assert_output_contains "wrote AGENTS.md and CLAUDE.md"
}

@test "config init leaves pre-existing AGENTS.md and CLAUDE.md untouched" {
    create_global_config
    echo "hand-written AGENTS.md" > "$TEST_WORK/AGENTS.md"
    echo "hand-written CLAUDE.md" > "$TEST_WORK/CLAUDE.md"
    run "$STQRY" config init --name=testsite
    assert_success
    assert_output_contains "AGENTS.md and CLAUDE.md already exist, left untouched"
    run grep -c "hand-written AGENTS.md" "$TEST_WORK/AGENTS.md"
    [ "$output" = "1" ]
    run grep -c "hand-written CLAUDE.md" "$TEST_WORK/CLAUDE.md"
    [ "$output" = "1" ]
}

@test "config init writes only the missing doc" {
    create_global_config
    echo "hand-written CLAUDE.md" > "$TEST_WORK/CLAUDE.md"
    run "$STQRY" config init --name=testsite
    assert_success
    [ -f "$TEST_WORK/AGENTS.md" ]
    assert_output_contains "wrote AGENTS.md"
    assert_output_contains "CLAUDE.md already exists, left untouched"
    run grep -c "hand-written CLAUDE.md" "$TEST_WORK/CLAUDE.md"
    [ "$output" = "1" ]
}
