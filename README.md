### Panzer cf cli plugin

* customizable "cf a" output
* lookup route function, to find a route, it's domain and in which org and space it lives
* show audit events

**For "cf aa":**  
Choose the columns you want in your output with the envvar CF_COLS.  
Limit the output by specifying the appname prefix with the -a flag, only apps who's names start with that prefix will be shown.  
Instead of "cf apps" or "cf a" you now use **"cf aa [-a appname-prefix]"** to get the results.  

The environment variable **CF_COLS** can be used the specify a comma-separated list of column names.  
The following column names are supported (case sensitive): 

**Name,State,Memory,Disk,Type,#Inst,Host,Cpu%,MemUsed,Created,Updated,Buildpacks,HealthCheck,InvocTmout,Tmout,Guid,ProcState,Uptime,InstancePorts**   

Mind that there are application related columns and application instance (process) related columns.  
From the above set of columns, the following are process-related: 

**Host, Cpu%, MemUsed, ProcState, Uptime, InstancePorts**.  


If you specify one ore more of these columns, you will get data for each instance of an app. Specifying one of these columns makes the command slower, especially if the space has many apps. (one cf API call per app is required, like the regular "cf apps" command does.)

To get all columns (you need a wide screen), specify: **CF_COLS=ALL**

**For "cf lr":**  
You specify the hostname using the -r flag "cf lr -r my-test-app", and it will search the route(s) and the domains and in which org and space they live and present it in a table.  
If you specify the -t flag you will also be cf targeted to the org/space where the route was found.

**For "cf ev":**  
You can filter the output by optionally specifying one or more of the following flags (the value specified should be contained in the given field):  
* -l (--limit) - the maximum numbers of results to show (currently limited to 5000)
* -a (--action) - client side filter by action (i.e. delete)
* -t (--target) - client side filter by target (i.e. myAppName)
* -y (--type) - client side filter by type (i.e. app or service_instance)
* -c (--actor) - client side filter by actor (i.e. user4711)
* -o (--org) - server side filter by org name
* -s (--space) - server side filter by space name

An example to use all filters:  `cf ev --limit 4381 --action delete --target testapp --type route --actor user4711 --org my-org --space my-space`

Install the plugins as usual with _cf install-plugin <plugin binary>_
