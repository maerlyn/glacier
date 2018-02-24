# glacier

An aws/glacier client intended for my personal use, so you can say it's features are somewhat limited.

So far it can:

 - list your vaults
 - get a vault inventory, it waits for the job to complete the outputs the json response
 - upload a file, both small and large (separator is 4GB)
