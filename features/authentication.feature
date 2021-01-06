#noinspection CucumberUndefinedStep
Feature: Authentication

  After setting up the SSH and authentication servers we should be able to authenticate via SSH with the correct users
  and fail with the incorrect ones.

  Scenario: Authentication should fail with invalid user
    When I delete user "foo"
    Then authentication with user "foo" and password "bar" should fail

  Scenario: Authentication should fail with wrong password
    When I create the user "foo" with the password "bar"
    Then authentication with user "foo" and password "baz" should fail

  Scenario: Authentication should succeed with correct password
    When I create the user "foo" with the password "bar"
    Then authentication with user "foo" and password "bar" should succeed
