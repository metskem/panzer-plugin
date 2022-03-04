### Panzer cf cli plugin

A plugin for faster interaction (less API calls) with Cloud Foundry.
Choose the columns you want in your output with the envvar CF_COLS.  
Limit the output by specifying the appname prefix, only apps who's names start with that prefix will be shown.  
Instead of "cf apps" or "cf a" you now use **"cf aa"** to get the results.  

The environment variable **CF_COLS** can be used the specify a comma-separated list of column names.  
The following column names are supported: 

**Name,State,Memory,Disk,Type,#Inst,Host,Cpu%,MemUsed,Created,Updated,Buildpacks,HealthCheck,InvocTmout,Tmout,Guid,ProcState,Uptime,InstancePorts**   

Mind that there are application related columns and application instance (process) related columns.  
From the above set of columns, the following are process-related: 

**Host, Cpu%, MemUsed, ProcState, Uptime, InstancePorts**.  


If you specify one ore more of these columns, you will get data for each instance of an app. Specifying one of these columns makes the command slower, especially if the space has many apps. (one cf API call per app is required, like the regular "cf apps" command does.)

To get all columns (you need a wide screen), specify: **CF_COLS=ALL**

Install the plugins as usual with _cf install-plugin <plugin binary>_
