## Testing

* Ensure that every code change is accompanied by corresponding updates to
  the test suite. If the modified function or component does not already have
  tests, create them as part of the same change.
* Tests must cover successful operation as well as all execution paths,
  logical branches, meaningful edge cases, and failure conditions.
  Validate logical invariants and expected behavior, and also confirm
  correct handling of invalid inputs and failure conditions.
* When working with text processing, parsing, or regular expressions,
  design tests that explicitly validate each capture group, branch condition,
  and alternative match path so that all parsing logic is exercised.
* No new or modified behavior should remain untested.
