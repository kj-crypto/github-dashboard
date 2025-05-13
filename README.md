# GitHub Repository Dashboard

A terminal-based dashboard for monitoring your GitHub repositories and contributions.

## Features

- View your GitHub repositories in a sortable table
- See your GitHub contribution calendar
- Browse repository READMEs directly in the terminal
- Lightweight and fast terminal interface

## Installation

### Using Go
```
go install github.com/kj-crypto/github-dashboard@latest 
```

### Build from sources
```
git clone https:// github.com/kj-crypto/github-dashboard.git 
cd github-dashboard
go build -o github-dashboard ./cmd/cli
```

## Usage
First setup Github token
```
export GITHUB_TOKEN=your_github_token
````

Run `github-dashboard <username>`

### Navigation
 - `↑/↓`: navigate repositories
 - `→`: enter readme section scrolling
 - `←`: back to repos list scrolling


## License
MIT
