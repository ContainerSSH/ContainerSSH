Feature: Command execution
  Scenario: Simple command execution via Docker
    Given I started the SSH server
    And I started the authentication server
    And I started the configuration server
    And I created the user "foo" with the password "bar"
    And I configure the user "foo" to use Docker
    Then I should be able to execute a command with user "foo" and password "bar"

  Scenario: Simple command execution via Kubernetes
    Given I started the SSH server
    And I started the authentication server
    And I started the configuration server
    And I created the user "foo" with the password "bar"
    And I configure the user "foo" to use Kubernetes
    Then I should be able to execute a command with user "foo" and password "bar"
