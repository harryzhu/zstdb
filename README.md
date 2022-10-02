# sqlconf

## Usage
### in go.mod
<code>
    replace sqlconf => ../sqlconf
</code>

### in go file
<code>
    import "sqlconf"

    var config *sqlconf.Conf = new(sqlconf.Conf)

    config.Open("./conf.db").Refresh().Print()

    config.Set("KEY", "VALUE")
</code>

