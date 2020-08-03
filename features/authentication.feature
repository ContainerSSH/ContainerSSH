Feature: Authentication
  Scenario: Authentication should fail with non-existent users
    Given I started the SSH server
    And I started the authentication server
    And I started the configuration server
    Then authentication with user "foo" and password "bar" should fail

  Scenario: Authentication should fail with wrong password
    Given I started the SSH server
    And I started the authentication server
    And I started the configuration server
    And I created the user "foo" with the password "bar"
    Then authentication with user "foo" and password "baz" should fail

  Scenario: Authentication should succeed with correct password
    Given I started the SSH server
    And I started the authentication server
    And I started the configuration server
    And I created the user "foo" with the password "bar"
    Then authentication with user "foo" and password "bar" should succeed
