# Case Tracking Example

This example demonstrates using `agentops` for file-driven case lifecycle management.

## Setup

```bash
cd examples/case-tracking
agentops doctor
```

## Usage

```bash
agentops case create fix-login-bug
agentops case list
agentops case transition CASE-20260321-fix-login-bug start
agentops case list --status active
```
