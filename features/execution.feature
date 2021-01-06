#noinspection CucumberUndefinedStep
Feature: Execution

  After setting up ContainerSSH we should be able to execute programs.

  Scenario: Command execution should work and the output should be returned.
    Given I created the user "moby" with the password "bar"
    When I open an SSH connection with the user "moby" and the password "bar"
    And I open an SSH session
    And I set the environment variable "MESSAGE" to the value "Hello world!"
    And I execute the command "echo \"$MESSAGE\""
    Then I should see "Hello world!" in the output
    And the session should exit with the code "0"
