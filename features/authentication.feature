Feature: Authentication
  Scenario:
    Given I started the SSH server
    And I started the authentication server
    Then Authentication with user "foo" and password "bar" should fail
