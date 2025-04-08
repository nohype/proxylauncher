# ProxyLauncher

A tiny, zero-dependency Windows application that passes its command line arguments on to a executable, either prepending or appending additional arguments specified in a configuration file.

## Overview

ProxyLauncher serves as a wrapper for other executables, allowing you to add predefined arguments to the target application without modifying the original command line. This is particularly useful when you want to add command line arguments but have no control over its invocation, or you want to create a drop-in replacement for an existing executable which needs additional arguments to function.

## Configuration

Create a configuration file with the same name as the executable but with a `.cfg` extension in the same directory. For example, if your executable is named `proxylauncher.exe`, the configuration file should be named `proxylauncher.cfg`.

The configuration file uses a simple key-value format:

```
target=path\to\target\application.exe
extraArgs=--arg1 --arg2 "argument with spaces" /switch=value
extraArgsOrder=before
```

### Configuration Keys

- `target`: Path to the target executable (absolute or relative to the ProxyLauncher's directory)
- `extraArgs`: Additional arguments to pass to the target executable
- `extraArgsOrder`: Determines whether the extra arguments are added before or after the command line arguments (valid values: `before` or `after`)

## Usage

### Simple Usage

Simply run the ProxyLauncher executable with any arguments you want to pass to the target application:

```
proxylauncher.exe arg1 arg2 arg3
```

Based on your configuration, ProxyLauncher will execute the target application with the combined arguments.

### Common Usage
Replace an existing executable with ProxyLauncher:

1. Rename the original executable (e.g. `target_app.exe`)to something else (e.g., `original.exe`), or move it to another directory (incl. all its dependencies, like DLLs)
2. Copy `proxylauncher.exe` to the same directory and rename it to the original executable's name (e.g., `target_app.exe`)
3. Create a configuration file with the same name as the executable but with a `.cfg` extension (e.g., `target_app.cfg`)
4. Edit the configuration file to set the `target` to the original executable's path
5. Edit the configuration file to set the `extraArgs` to the arguments you want to pass to the target application
6. Edit the configuration file to set the `extraArgsOrder` to `before` or `after` based on your preference

When the `target_app.exe` executable is now invoked, it is the ProxyLauncher that is executed, which will pass the arguments on to the original target application and prepend or append the configured additional arguments.

## Bonus
When launching without an existing configuration file, ProxyLauncher will create a default configuration file and open it in Notepad. You'll just need to fill in the values.

## Building from Source

To build a static Windows executable with zero runtime dependencies:

```
go build -ldflags="-H windowsgui" -o proxylauncher.exe
```

The `-H windowsgui` flag creates a Windows GUI application (no console window).

## Example

If your configuration file is `proxylauncher.cfg` and contains:
```
target=.\child directory\another_program.exe
extraArgs=--specialMode /switch=ON
extraArgsOrder=before
```

And you run, with "C:\" being the working directory:
```
proxylauncher.exe A B C=D /E=OFF -nogood
```

The proxy launcher will execute:
```
"C:\path\to\child directory\another_program.exe" --specialMode /switch=ON A B C=D /E=OFF -nogood
```
