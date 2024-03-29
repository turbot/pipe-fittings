# pipes-fittings

Shared Pipes Component

## v1.0.3 [2023-03-26]

_Bug fixes_

* Enable `loop` block for `container`, `function`, `message` and `input` steps.
* Allow using HCL expression for `max_currency` attribute.
* `throw`, `error` and `retry` block now works for `input` step.

## v1.0.2 [2023-03-18]

_Bug fixes_

* Add resource metadata after loading mod definition. ([#372](https://github.com/turbot/pipe-fittings/issues/372)).

## v1.0.1 [2023-03-15]

_Bug fixes_

* Erroneous error message detecting a missing credential where there isn't one.
* HCL `try()` function should be evaluated at runtime rather than parse time.
* Add filename and line number information in step validation error messages.

## v1.0.0 [2023-03-14]

_What's new?_

* Optimize workspace load time for large workspaces, i.e. multiple dependent mods.
* Strip quotes in a string if it exists in the beginning and end of the string for string -> type coerce function used by Flowpipe to parse CLI args.
* Better error messages for Flowpipe pipeline run (just for credentials currently).

## v0.3.4 [2024-03-08]

_Bug fixes_

* Only detect logical changes in step's throw block.
* HCL expression comparison for conditional operator now works.

## v0.3.3 [2024-03-07]

_What's new?_

* Optimize workspace load time when only variables are being loaded.  ([#357](https://github.com/turbot/pipe-fittings/issues/357)). 

## v0.3.2 [2024-03-07]

_Bug fixes_

* Pipeline step input should not trigger a re-load with empty line changes in its definition. ([#297](https://github.com/turbot/pipe-fittings/issues/297)).

## v0.3.1 [2024-03-06]

_What's new?_

* Support for Powerpipe workspace profiles.

_Bug fixes_

* Better error message for variable validation errors to indicate the variable location if available. ([#356](https://github.com/turbot/pipe-fittings/issues/356)).
* Add support for loading variables from  legacy (steampipe) vars file. ([#350](https://github.com/turbot/pipe-fittings/issues/350)).
* Add missing snapshot tags for dashboard resources. ([#355](https://github.com/turbot/pipe-fittings/issues/355)).

## v0.3.0 [2024-03-05]

_What's new?_

* Credential Import, Notifier and Integration resources.

_Bug fixes_

* Triggers are missing required field validation ([#225](https://github.com/turbot/pipe-fittings/issues/255)).
* Missing pipeline output validation ([#239](https://github.com/turbot/pipe-fittings/issues/239)).
* Validate default param value against param type ([#262]((https://github.com/turbot/pipe-fittings/issues/262))
* Removed titles when merging multiple error messages ([#263]((https://github.com/turbot/pipe-fittings/issues/263))
* File watcher reliability improvements.

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