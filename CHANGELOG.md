# pipes-fittings

Shared Pipes Component

## v2.1.1 [2025-02-11]
* Get credentials from the platform-default native credentials store for turbot-hosted private plugins. ([#643](https://github.com/turbot/pipe-fittings/issues/643))

## v2.1.0 [2025-02-11]
* TailpipeConnection now supports indexes/partitions in addition to to/from. ([#632](https://github.com/turbot/pipe-fittings/issues/632)).
* Move `secrets` package from steampipe-plugin-code.

## v2.0.1 [2025-02-02]

* Add backend.NameFromConnectionString. ([#631](https://github.com/turbot/pipe-fittings/issues/631)).
* Show row count only when paged in querydisplay table
* Add constants.ArgQuiet

## v2.0.0 [2025-01-30]

* Add support for `tailpipe` and tailpipe detections within `powerpipe`
* Remove flowpipe and powerpipe specific resources and move into their respective app repos. ([#603](https://github.com/turbot/pipe-fittings/issues/603)).
* ConnectionStringFromConnectionName returns ConnectionStringProvider
* DuckDBBackend installs and loads duckdb extensions from pipes extensions dir
* Add TailpipeConnection, which invokes `tailpipe connect`, passing time range
* Add TailpipeWorkspaceProfile
* Add support for decoding Detection and DetectionBenchmark
* Add reflect.InstanceOf to instantiate a an instance of a generic type
* Add IsValidDir() to determine if a given path is an existing directory
* Add SafeBoolEqual func
* Add tailpipe hub API endpoint
* Add hclhelpers.BlockRangeWithLabels
* Add Add EnsurePipesDuckDbExtensionsDir
* Add SqlFilter
* Add ParseTime
* Add UserConfirmationWithDefault
* Limit table query output to 10000 rows


## v1.7.3 [2024-01-07]

_What's new_

* Added support for installing mods from GitLab repositories.

## v1.7.2 [2024-12-23]

_What's new_

* Updated `golang.org/x/crypto` to v0.31.0.

## v1.7.1 [2024-12-23]

_What's new_

* Added `ResetCache` function to reset the all in-memory cache.

## v1.7.0 [2024-12-20]

_What's new_

* Improved performance of mod loading when there's a lot of `connections`.
* Added in memory cache functionality.

## v1.6.6 [2024-11-21]

_Bug fixes_

* Update uploaded pipes snapshot URL to include `/powerpipe`.

## v1.6.5 [2024-10-25]

_Bug fixes_

* Typos in error messages.
* Crashes in Go to HCL conversion for complex types.

## v1.6.4 [2024-10-24]

_Bug fixes_

* Coerce input variable entered from terminal to the correct type. ([#595](https://github.com/turbot/pipe-fittings/issues/595)).

## v1.6.3 [2024-10-23]

_Bug fixes_

* `form_url` should be sanitized.

## v1.6.2 [2024-10-22]

_Bug fixes_

* Removed buggy database connection string redaction that may cause some JSON string to be invalid. ([#594](https://github.com/turbot/pipe-fittings/issues/594)).

## v1.6.1 [2024-10-22]

_What's new_

 * Support setting a variable of type Connection using a connection string or cloud workspace handle. ([#592](https://github.com/turbot/pipe-fittings/issues/592)).


## v1.6.0 [2024-10-22]

_What's new_

* `Pipeling Connection` resource to replace `Credential`. ([#449](https://github.com/turbot/pipe-fittings/issues/449)).
* Custom type (`connection` and `notifier` ) for pipeline param and mod variable. ([#523](https://github.com/turbot/pipe-fittings/issues/523)).

_Bug fixes_

* Fix missing fields when importing Jira connections. ([#577](https://github.com/turbot/pipe-fittings/issues/577))

## v1.5.7 [2024-09-23]

_Bug fixes_

* Flowpipe pipeline will no longer crash if user provided a non-map type reference (another case).

## v1.5.6 [2024-09-23]

_Bug fixes_

* Flowpipe pipeline will no longer crash if user provided a non-map type reference.

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