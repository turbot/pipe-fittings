# pipes-fittings

Shared Pipes Component

## v1.6.0 [tbd]

_What's new_

* `Pipeling Connection` resource to replace `Credential`. ([#449](https://github.com/turbot/pipe-fittings/issues/449)).
* Convert credential for abuseipdb to connection. ([#453](https://github.com/turbot/pipe-fittings/issues/453)).
* Convert credential for alicloud to connection. ([#454](https://github.com/turbot/pipe-fittings/issues/454)).
* Convert credential for azure to connection. ([#456](https://github.com/turbot/pipe-fittings/issues/456)).
* Convert credential for bitbucket to connection. ([#457](https://github.com/turbot/pipe-fittings/issues/457)).
* Convert credential for clickup to connection. ([#458](https://github.com/turbot/pipe-fittings/issues/458)).
* Convert credential for datadog to connection, ([#459](https://github.com/turbot/pipe-fittings/issues/459)).
* Convert credential for discord to connection. ([#460](https://github.com/turbot/pipe-fittings/issues/460)).
* Convert credential for freshdesk to connection. ([#461](https://github.com/turbot/pipe-fittings/issues/461)).
* Convert credential for github to connection. ([#467](https://github.com/turbot/pipe-fittings/issues/467)).
* Convert credential for gcp to connection. ([#466](https://github.com/turbot/pipe-fittings/issues/466)).
* Convert credential for gitlab to connection. ([#468](https://github.com/turbot/pipe-fittings/issues/468)).
* Convert credential for ipstack to connection. ([#470](https://github.com/turbot/pipe-fittings/issues/470)).
* Convert credential for ip2locationio to connection. ([#469](https://github.com/turbot/pipe-fittings/issues/469)).
* Convert credential for jira to connection. ([#471](https://github.com/turbot/pipe-fittings/issues/471)).
* Convert credential for jumpcloud to connection. ([#472](https://github.com/turbot/pipe-fittings/issues/472)).
* Convert credential for mastodon to connection. ([#473](https://github.com/turbot/pipe-fittings/issues/473)).
* Convert credential for microsoft teams to connection. ([#474](https://github.com/turbot/pipe-fittings/issues/474)).
* Convert credential for okta to connection. ([#475](https://github.com/turbot/pipe-fittings/issues/475)).
* Convert credential for openai to connection. ([#476](https://github.com/turbot/pipe-fittings/issues/476)).
* Convert credential for opsgenie to connection. ([#477](https://github.com/turbot/pipe-fittings/issues/477)).
* Convert credential for pagerduty to connection. ([#478](https://github.com/turbot/pipe-fittings/issues/478)).
* Convert credential for sendgrid to connection. ([#479](https://github.com/turbot/pipe-fittings/issues/479)).
* Convert credential for servicenow to connection. ([#480](https://github.com/turbot/pipe-fittings/issues/480)).
* Convert credential for slack to connection. ([#481](https://github.com/turbot/pipe-fittings/issues/481)).
* Convert credential for trello to connection. ([#482](https://github.com/turbot/pipe-fittings/issues/482)).
* Convert credential for turbot guardrails to connection. ([#483](https://github.com/turbot/pipe-fittings/issues/483)).
* Convert credential for turbot pipes to connection. ([#484](https://github.com/turbot/pipe-fittings/issues/484)).
* Convert credential for uptime robot to connection. ([#485](https://github.com/turbot/pipe-fittings/issues/485)).
* Convert credential for urlscan to connection. ([#486](https://github.com/turbot/pipe-fittings/issues/486)).
* Convert credential for vault to connection. ([#486](https://github.com/turbot/pipe-fittings/issues/486)).
* Convert credential for virus total to connection. ([#488](https://github.com/turbot/pipe-fittings/issues/488)).
* Convert credential for zendesk to connection. ([#489](https://github.com/turbot/pipe-fittings/issues/489)).

_Bug fixes_

## v1.5.5 [2024-09-03]

_What's new_

* `tags` arguments to Flowpipe pipeline param and mod variable.

## v1.5.4 [2024-08-22]

_What's new_

* JSON extension support for `duckdb` backends. ([#442](https://github.com/turbot/pipe-fittings/issues/442))

## v1.5.3 [2024-08-22]

* Flowpipe triggers now appears as top level resources, allowing them to be listed from the root mod. ([#444](https://github.com/turbot/pipe-fittings/issues/444))

## v1.5.2 [2024-08-14]

* Flowpipe pipeline param default value compatibility test with the declared type may fail for complex types. ([#441](https://github.com/turbot/pipe-fittings/issues/441)) 

## v1.5.1 [2024-08-13]

_Bug fixes_

* CLI notification messages should be printed to stderr to avoid interfering with stdout. ([#437](https://github.com/turbot/pipe-fittings/issues/437))

## v1.5.0 [2024-08-13]

_What's new_

* `type_string` attribute to `Variable` and `PipelineParam` definition. `type` in its current format is deprecated and it will be changed native `cty type` JSON serialisation in the future. ([#435](https://github.com/turbot/pipe-fittings/issues/435)).
* `param` to `Trigger` definition.
* safe pointer dereference function.
* JSONTime type to handle time.Time in JSON output.
* `GoToHCLString` function that converts a Go data structure to an HCL string.

_Bug fixes_

* `CtyTypeToHclType` function to handle complex types.
* Pipeline param default value compatibility test with the declared type. ([#436](https://github.com/turbot/pipe-fittings/issues/436)).

## v1.4.3 [2024-07-12]

* Updated github.com/hashicorp/go-getter to v1.7.5.

## v1.4.2 [2024-07-03]

_Bug fixes_

* Unique column name generator should take hash using the column index as an input rather than appending occurrence index to the hashed output. ([#426](https://github.com/turbot/pipe-fittings/issues/426)).
* Fix exception when migrating steampipe mod lock file for powerpipe. ([#429](https://github.com/turbot/pipe-fittings/issues/429))

## v1.4.1 [2024-06-10]

* Update snapshot schema version to `20240607`. ([#423](https://github.com/turbot/pipe-fittings/issues/423))

## v1.4.0 [2024-06-07]

_What's new_

* Update mod install to only install or update mods which are command targets (and their dependencies). Default pull mode for install is `latest` if there is a target, and `minimal` if no target is given. ([#415](https://github.com/turbot/pipe-fittings/issues/415))
* Add UniqueNameGenerator functions to generate random unique column names for query JSON output. ([#417](https://github.com/turbot/pipe-fittings/issues/417))
* use `github.com/turbot/pipes-sdk-go` instead of `github.com/turbot/steampipe-cloud-sdk-go`. ([#418](https://github.com/turbot/pipe-fittings/issues/418))
* Updated `pgx` and `pgconn` to latest versions.

## v1.3.4 [2024-05-31]

_Bug fixes_

* Params should be ordered as defined in the pipeline definition ([#408](https://github.com/turbot/pipe-fittings/issues/408))

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