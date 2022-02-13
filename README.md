### Panzer cf cli plugin

A plugin for faster interaction (less API calls) with Cloud Foundry, and choose the columns you want in your output.  
Instead of "cf apps" or "cf a" you now use **"cf aa"** to get the results.  
The environment variable **CF_COLS** can be used the specify a comma-separated list of column names.  
The following column names are supported: **Name,State,Memory,Disk,Type,#Inst,Host,Cpu,MemUsed,Created,Updated,Buildpacks,HealthCheck,InvocTmout,Tmout,Guid,ProcState,Uptime,InstancePorts**   
Mind that there application related columns and application instance (process) related columns.  
From the above set of columns, the following are process-related: **Host, Cpu, MemUsed, ProcState, Uptime, InstancePorts**.  
If you specify one ore more of these columns, you will get data for each instance of an app. Also mind that specifying one of these columns makes the command a lot slower, especially if the space has many apps. (one cf API call per app is required, like the regular "cf apps" command does.)

Install the plugins as usual with _cf install-plugin <plugin binary>_
