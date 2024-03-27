# Use better HTTP status codes in API

* Status: accepted
* Deciders: @artem-smotrakov, @zuBux
* Date: 2021-12-09

## Context and Problem Statement

The KeyConjurer APIs return 200 whether an error occurred or not. It may return 5xx if a fatal error occurred during Lambda execution.
Instead, the API should return proper HTTP status codes:
* 2xx in case of success
* 4xx in case of user's mistake
* 5xx in case of server failure

## Considered Options

* **Option 1:** Set HTTP status codes in the Lambda code and use the AWS API Gateway proxy integration.
* **Option 2:** Set HTTP status codes in the AWS API Gateway by matching error strings with Regex and applying Velocity templates to the Lambda's error responses.

## Decision Outcome

Chosen **Option 1**, because it keeps Terraform configs simpler,
doesn't introduce Velocity Template Language to the project,
and it is easier to cover with unit tests.

## Pros and Cons of the Options

### Option 1

* Good, because it keeps Terraform configs simpler.
* Good, because it is easier to cover with unit tests.
* Bad, because it makes the Lambda handlers a bit more complex.

### Option 2

* Good, because it keeps the Lambda handler's code simpler.
* Bad, because it makes Terraform configs more complex.
* Bad, because it introduces Velocity Template Language to the project.
* Bad, because it is difficult to test automatically in CI/CD.
