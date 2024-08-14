# Juno-plugin
This repository provides an example plugin for Juno.

# Todo:

Remove the replace command in go.mod when plugin is supported on Junos main branch

# Important:

Go plugings require that both the application (Juno in this case) and the plugin (myplugin.go) are built with the exact same version of Go, and, the same dependecies. 

# Workflow:

1. Write your code in myplugin.go
2. Run 'make build'. This should generate an '.so' file, which you will need to pass into Juno.
3. Download [Juno](https://github.com/NethermindEth/juno) (in a seperate directory), and run it (eg 'make juno').
4. Pass the '.so' file generated in step 2 above, into Juno. For example, './build/juno --plugin-path ./path/to/myplugin.so'
