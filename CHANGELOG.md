# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.2] - 2020-04-01

### Added
- Support non-string types in query and header parameters.
- Support additionalProperties when no properties are defined, creating maps.
- Support allOf. Referenced schemas are included as embedded structs. Inline
  schemas have their fields included in the main struct directly.

### Fixed
- Support references in responses.
- Fixed an issue where reserved keywords were being used as function/method
  arguments.
- Fix support for reference parameters.
- Build on the latest version of Go (1.14.x)

## [0.0.1] - 2017-10-15

### Added
- Initial release of `oag`.
