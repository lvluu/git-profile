# Git Profile CLI

## Overview
`git-profile` is a simple CLI tool to manage multiple Git profiles easily.

## Installation

### Go Install
```bash
go install github.com/yourusername/git-profile@latest
```

### Homebrew (macOS)
```bash
brew tap yourusername/git-profile
brew install git-profile
```

### Manual Installation
Download the appropriate binary from [Releases](https://github.com/yourusername/git-profile/releases)

## Usage

### List Profiles
```bash
git profile ls
```

### Add a Profile
```bash
git profile add work "John Doe" john.doe@company.com
```

### Apply a Profile
```bash
git profile apply work
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first.

## License
[MIT](https://choosealicense.com/licenses/mit/)