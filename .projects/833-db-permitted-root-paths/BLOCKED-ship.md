## Kind: design
Question: When the permitted-folders setting is on, should a host's DuckDB startup script still be able to install and load extensions, such as tailpipe's connect script does?
Research done: Probed the branch with the setting set. DuckDB is locked, so its init script fails at `install` or `load` with a permission error, even with the script's folder and the extension folder both listed. The issue says listing the folders is how a host keeps its own DuckDB files working, and this case breaks that. Checked: the init script runs before the lock, so the fix that would be possible here is to load the extensions a host names before the lock; that needs tailpipe's script to change too.
Options: Load the extensions the host names before the lock, and tailpipe's script drops its install lines; or declare tailpipe unsupported while the setting is on, so its connections are refused with a clear message.
Recommendation: Load the named extensions before the lock, because it keeps tailpipe working under the setting and the issue's stated promise holds. It needs a change in tailpipe too, so the ship waits for that.
If nobody answers: The ship stays held at this blocker. Nothing is opened as a PR.
Stops for: product decision
