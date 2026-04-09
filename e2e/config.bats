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
