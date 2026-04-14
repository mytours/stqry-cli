#!/usr/bin/env bats
# Requires cassettes recorded via: make record TEST_API_URL=https://api-us.stqry.com
# Without cassettes the replay proxy returns 404 and these tests fail — that is expected.
load "test_helper"

@test "screens list shows screens" {
    create_global_config
    run "$STQRY" screens list --site=testsite --json
    assert_success
}

@test "screens list --quiet outputs JSON array" {
    create_global_config
    run "$STQRY" screens list --site=testsite --quiet
    assert_success
}
