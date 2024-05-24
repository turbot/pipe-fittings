# pipes-fittings

Shared Pipes Component

## v1.3.3 [2024-05-23]

_Bug fixes_

* Fix mod `require` block being rewritten incorrectly when installing a mod if the require block exists but does not contain mod requirements. ([#406](https://github.com/turbot/pipe-fittings/issues/406))
* Fix install status display when  updating transitive dependencies. 

## v1.3.2 [2024-05-23]

_Bug fixes_

* Apps should respect the app version constraint defined in the correctly name `require` sub-block. ([#405](https://github.com/turbot/pipe-fittings/issues/405))

## v1.3.1 [2024-05-21]

_Bug fixes_

* Unresolved pipeline should be added to the "unresolved block" so it can be resolved in a subsequent parse. ([#402](https://github.com/turbot/pipe-fittings/issues/402))

## v1.3.0 [2024-05-13]


## v1.2.2 [2024-05-11]

_Bug fixes_

* File load ordering issue with map type `locals`. Manifest in issue ([#399](https://github.com/turbot/pipe-fittings/issues/399))


## v1.2.1 [2024-05-10]

_Bug fixes_

* Trigger's common attributes (title, description, tags, documentation) should allow function and expresion. ([#394](https://github.com/turbot/pipe-fittings/issues/394)).

## v1.2.0 [2024-04-23]

_What's new?_
 
- Add `benchmark_timeout` and `dashboard_timeout` to Powerpipe workspace profile ([#391](https://github.com/turbot/pipe-fittings/issues/391)).

## v1.1.2 [2024-04-22]

_Bug fixes_
 
- When calling mod update, respect the argument (if any) and only update specified mods. ([#388](https://github.com/turbot/pipe-fittings/issues/388)).
- Fix display of updates to transitive dependencies. ([#389](https://github.com/turbot/pipe-fittings/issues/389)).

## v1.1.1 [2024-04-10]

_Bug fixes_

- Update variable parsing to better handle extraneous space characters - update sanitiseVariableNames to handle multiple spaces. ([#384](https://github.com/turbot/pipe-fittings/issues/384)).

## v1.1.0 [2024-04-09]

_What's new?_

* Support installing private mods using a github app token. ([#381](https://github.com/turbot/pipe-fittings/issues/381)).
* Update mod installation to use app-specific git token env vars - POWERPIPE_GIT_TOKEN and FLOWPIPE_GIT_TOKEN. ([#382](https://github.com/turbot/pipe-fittings/issues/382)).

## v1.0.5 [2024-04-12]

_Bug fixes_

* Fixed loop equality check.
* Duplicate step names are now detected and reported as an error.
* Crash when using param in query step's args.
* Better error message for invalid notifier reference.

## v1.0.4 [2024-04-01]

_Bug fixes_

* Fixed misleading error messsage when parsing step dependencies.

## v1.0.3 [2024-03-26]

_Bug fixes_

* Enable `loop` block for `container`, `function`, `message` and `input` steps.
* Allow using HCL expression for `max_currency` attribute.
* `throw`, `error` and `retry` block now works for `input` step.

## v1.0.2 [2024-03-18]

_Bug fixes_

* Add resource metadata after loading mod definition. ([#372](https://github.com/turbot/pipe-fittings/issues/372)).

## v1.0.1 [2024-03-15]

_Bug fixes_

* Erroneous error message detecting a missing credential where there isn't one.
* HCL `try()` function should be evaluated at runtime rather than parse time.
* Add filename and line number information in step validation error messages.

## v1.0.0 [2024-03-14]

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