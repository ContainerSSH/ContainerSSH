#noinspection CucumberUndefinedStep
Feature: Command execution

  After setting up the SSH and the authentication and configuration servers we should be able to execute a command
  via SSH and return the result.

  Scenario: Simple command execution via Docker
    Given I started the SSH server
    And I started the authentication server
    And I started the configuration server
    And I created the user "moby" with the password "bar"
    And I configure the user "moby" to use Docker
    Then I should be able to execute a command with user "moby" and password "bar"

  Scenario: Simple command execution via Kubernetes
    Given I started the SSH server
    And I started the authentication server
    And I started the configuration server
    And I created the user "foo" with the password "bar"
    And I configure the user "foo" to use Kubernetes
    Then I should be able to execute a command with user "foo" and password "bar"
