### Panzer cf cli plugin

* customizable "cf a" output
* lookup route function, to find a route, it's domain and in which org and space it lives
* show audit events

**For "cf aa":**  
Choose the columns you want in your output with the envvar CF_COLS.  
Limit the output by specifying the appname prefix with the -a flag, only apps who's names start with that prefix will be shown.  
Instead of "cf apps" or "cf a" you now use **"cf aa [-a appname]"** to get the results (appname is a regular expression).  
Use the -q (--hide-headers) to hide the column headers and the summary at the bottom (handy for processing the output).

The environment variable **CF_COLS** can be used the specify a comma-separated list of column names.  
The following column names are supported (case sensitive): 

**Name,State,Memory,LogRate,Disk,Type,#Inst,Host,Cpu%,MemUsed,LogRateUsed,Created,Updated,Buildpacks,Stack,HealthCheck,InvocTmout,Tmout,Guid,ProcState,ProcType,Uptime,InstancePorts**   

Mind that there are application related columns and application instance (process) related columns.  
From the above set of columns, the following are process-related: 

**Host, Cpu%, MemUsed, LogUsed, ProcState, Uptime, InstancePorts**.  


If you specify one ore more of these columns, you will get data for each instance of an app. Specifying one of these columns makes the command slower, especially if the space has many apps. (one cf API call per app is required, like the regular "cf apps" command does.)

To get all columns (you need a wide screen), specify: **CF_COLS=ALL**

**For "cf lr":**  
You specify the hostname using the -r flag "cf lr -r my-test-app", and it will search the route(s) and the domains and in which org and space they live and present it in a table.  
If you specify the -t flag you will also be cf targeted to the org/space where the route was found.

**For "cf ev":**  
You can filter the output by optionally specifying one or more of the following flags:

    -h --help          Displays help with available flag, subcommand, and positional value parameters.
    -l --limit         Limit the output to max XXX events (default: 500)
    -e --event-type    Filter the output (server side), (comma separated list of) event type to exactly match the filter (i.e. audit.app.update,app.crash)
    -n --target-name   Filter the output (client side), target name to fuzzy match the filter
    -t --target-type   Filter the output (client side), target type to fuzzy match the filter (i.e. app service_binding route)
    -a --actor         Filter the output (client side), actor name to fuzzy match the filter
    -o --org           Filter the output (server side), org name to exactly match the filter
    -s --space         Filter the output (server side), space name to exactly match the filter

    -q --hide-headers  Hide the column headers (handy for processing the output).

An example to use all filters:  `cf ev --limit 4381 --event-type audit.app.stop --target-name testapp --target-type route --actor user4711 --org my-org --space my-space`

**Installation and upgrade**
Download latest version from [releases](https://github.com/metskem/panzer-plugin/releases/latest)

Install the plugins as usual with 
```
cf install-plugin <plugin binary>
```
