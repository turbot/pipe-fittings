# pipes-fittings

Shared Pipes Component

## v0.3.0 [tbd]

_What's new?_

* Credential Import, Notifier and Integration resources.

_Bug fixes_

* Triggers are missing required field validation ([#225](https://github.com/turbot/pipe-fittings/issues/255)).
* Missing pipeline output validation ([#239](https://github.com/turbot/pipe-fittings/issues/239)).
* Validate default param value against param type ([#262]((https://github.com/turbot/pipe-fittings/issues/262))
* Removed titles when merging multiple error messages ([#263]((https://github.com/turbot/pipe-fittings/issues/263))

## v0.2.2 [2024-02-02]

_Bug fixes_

* Missing error handling during the conversion of Go struct to CTY value.
* Invalid `type` in pipeline param definition should throw a parsing error ([#252](https://github.com/turbot/pipe-fittings/issues/252)).

## v0.2.1 [2024-02-02]

_Bug fixes_

* Incorrect conversion of Go struct to CTY value.

## v0.2.0 [2024-01-24]

_What's new?_

* Added credentials support for the following plugins: 
  - BitBucket
  - Datadog
  - Freshdesk
  - JumpCloud
  - ServiceNow 
  - Turbot Guardrails
* Added Query Trigger mod resource

_Enhancements_

* Container step now supports `Source` in addition to `Image`.
* Added `enabled` attribute to Flowpipe Triggers.
* Added `method` block to HTTP Trigger
* New intervals (`5m`, `10m`, `15m`, `30m`, `60m`, `1h`, `2h`, `4h`, `6h`, `12h`, `24h`) are now supported for the Schedule and Query Triggers.

_Bug fixes_

* Fixed an issue in the bootstrap process for identifying the config path.

## v0.1.0 [2023-12-13]

Shared components for use across pipe projects.