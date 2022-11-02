<!-- New changes go on top. But below these comments. -->

<!-- We can track changes from Git logs, so why we needs this file? 

Guiding Principles
- Changelogs are for humans, not machines.
- There should be an entry for every single version.
- The same types of changes should be grouped.
- Versions and sections should be linkable.
- The latest version comes first.
- The release date of each version is displayed.
- Mention whether you follow Semantic Versioning.

Types of changes
- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Deprecated` for soon-to-be removed features.
- `Removed` for now removed features.
- `Fixed` for any bug fixes.
- `Security` in case of vulnerabilities.

-->

## [0.4.3](#043)

### Fixed
- Fix 500 response on `/vm/*` endpoints.

### Removed
- Remove internal endpoints.

### Changed
- Refactor `vm` module for speed and readability.

## [0.4.2](#042)

### Added
- Trigger docker deployment 

### Removed
- Don't remove image after GPU load is run.

### Fixed
- Fix one logical error where DMS on compute provider sent error ack for expired message. 
