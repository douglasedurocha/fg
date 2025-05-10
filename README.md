# FG - Java Application Manager

FG is a simple CLI tool that helps you manage different versions of Java applications. It can download, install, and run Java applications, while handling JDK dependencies automatically.

## Features

- Download and install specific versions of a Java application
- Manage JDK dependencies
- Install Maven dependencies
- Start and stop application instances
- Monitor running instances
- View application logs
- Uninstall versions when no longer needed

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/douglasedurocha/fg.git
cd fg

# Build the executable
go build -o fg cmd/fg/main.go

# Move the executable to a directory in your PATH (optional)
sudo mv fg /usr/local/bin/
```

## Usage

### List installed versions

```bash
fg list
```

### List available versions

```bash
fg available
```

Example output:
```
Version     Release Date
--------    ------------
2.0.0       2024-04-05
1.2.0       2024-03-10
1.1.0       2024-02-22
1.0.0       2024-01-15
```

### Install a specific version

```bash
fg install 1.0.0
```

### Install the latest version

```bash
fg update
```

### View configuration for a version

```bash
fg config 1.0.0
```

### Start an application

```bash
# Start with a specific version
fg start 1.0.0

# Start with the latest installed version
fg start
```

### Check running instances

```bash
fg status
```

### View logs

```bash
# View logs for a specific instance
fg logs 1234  # where 1234 is the PID

# View logs for the most recent instance
fg logs
```

### Stop an instance

```bash
# Stop a specific instance
fg stop 1234  # where 1234 is the PID

# Stop all running instances
fg stop
```

### Uninstall a version

```bash
fg uninstall 1.0.0
```

## Configuration

FG stores all data in the `~/.fg` directory:

- `~/.fg/versions/` - Installed versions
- `~/.fg/jdk/` - JDK installations
- `~/.fg/logs/` - Application logs
- `~/.fg/downloads/` - Temporary download files

## License

MIT 