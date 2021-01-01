#noinspection CucumberUndefinedStep
Feature: Command execution

  After starting the SSH and metrics server the documented metrics should be visible in the output.

  Scenario: Metrics visibility
    Given I started the SSH server
    Then the "containerssh_auth_server_failures" metric should be visible
    And the "containerssh_auth_success" metric should be visible
    And the "containerssh_auth_failures" metric should be visible
    And the "containerssh_config_server_failures" metric should be visible
    And the "containerssh_ssh_connections" metric should be visible
    And the "containerssh_ssh_handshake_successful" metric should be visible
    And the "containerssh_ssh_handshake_failed" metric should be visible
    And the "containerssh_ssh_current_connections" metric should be visible
