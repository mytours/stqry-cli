#!/usr/bin/env bats
load "test_helper"

@test "collections list shows collections" {
    create_global_config
    run "$STQRY" collections list --site=testsite --json
    assert_success
}

@test "collections list --quiet outputs JSON array" {
    create_global_config
    run "$STQRY" collections list --site=testsite --quiet
    assert_success
}
