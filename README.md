# Juno-plugin
This repository provides an example plugin for Juno.

# Todo:

Remove the replace command in go.mod when plugin is supported on Junos main branch

# Important:

Go plugings require that both the application (Juno in this case) and the plugin (myplugin.go) are built with the exact same version of Go, and, the same dependecies. 

# How to build

1. Fetch the Juno submodule: `git submodule update --init --recursive`
2. Build Juno: `cd juno && make juno`
3. Run `make build` from the root of this repository. This should generate an `.so` file, which you will need to pass into Juno.
4. Pass the `.so` file generated in step 2 above, into Juno. For example, `./build/juno --plugin-path ./path/to/myplugin.so`

