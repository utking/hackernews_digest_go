## Go version of HackerNews Digest

### Requirements

* Sqlite 3+

### Configuration

There is a default config-file name - `config.json`, but it can be overwritten by setting an ENV variable `CONFIG_FILENAME`. Note that it can be an absolute or a relative path.

To create a config-file, copy `config.example.json` to `config.json` (or any other name that seems right for you) and adjust what you think should be adjusted.

#### Output to console

Set "EmailTo" to an empty string if you don't want to send emails but simply want to print out the digest to the console. Setting "EmailTo" to a non-empty string but having "Smtp.Host" empty, you prevent any output.
