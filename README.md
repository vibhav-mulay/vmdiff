# File diffing and updating Tool

File diffing and updating tool based on rdiff, but using newer algorithms for generating chunks.
The tool currently supports FastCDC and Rabin-fingerprint based Content-Defined Chunking algorithms.
It uses Protocol Buffers (protobuf) to efficient encode the data when writing to files.

## Usage
```
File diffing tool

Usage:
  vmdiff [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  delta       Generate delta file.
  help        Help about any command
  patch       Patch old file with delta to generate new file.
  signature   Generate signature file.

Flags:
  -h, --help      help for vmdiff
  -v, --verbose   Verbose
```

The main actions are:
1. signature: Creates a signature of the input file, describing chunk information like hash, size and offset.
2. delta: Creates delta files describing changes between two files. The older file is input in terms of its signature.
3. patch: Applies the delta to the input file and creates a new files having all the updates.


## Build
The tool can be built using the `make`
```
# Build the tool along with all the required dependent files (protobuf)
make

# Install the tool
make install

# Run unit tests
make test

# Run unit tests with coverage
make coverage

# Clean all generated files
make clean
```
