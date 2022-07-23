## Go version of HackerNews Digest

### Requirements

* Sqlite 3+ OR MySQL 5.7+

### Configuration

There is a default config-file name - `config.json`. Note that it can be overwritten in the comman line (-c|--config). The path can be relative or absolute.

To create a config file, copy `config.example.json` to `config.json` (or any other name that seems right for you) and adjust what you think should be adjusted. Database driver can be `sqlite3` or `mysql`. For `sqlite3`, the Database string will be the name of the DB file.

#### Output to console

Set "EmailTo" to an empty string if you don't want to send emails but simply want to print out the digest to the console. Setting "EmailTo" to a non-empty string but having "Smtp.Host" empty, you prevent any output.

### Arguments

* -r|--reverse - to reverse the filtering
* -v|--vacuum - to remove old records, without running news updates (retention period is set set in the config file)
* -c|--config - to set a config file
