# JSPyrate

JSPyrate is an efficient and powerful tool designed to analyze JavaScript files, identify endpoints, URLs, and hardcoded secrets. With its multi-threading capabilities, JSPyrate can process multiple JavaScript files simultaneously, making it perfect for security researchers, bug bounty hunters, and developers alike.

## Features

- Downloads JavaScript files from a provided list of URLs
- Analyzes JavaScript files using multiple threads for faster processing
- Extracts endpoints and URLs from JavaScript files
- Searches for hardcoded secrets using a customizable wordlist of regex patterns
- Generates an organized output for each analyzed JavaScript file

## Installation

1. Clone the repository
   ```
   git clone https://github.com/yourusername/JSPyrate.git
   ```
2. Change to the JSPyrate directory
   ```
   cd JSPyrate
   ```
3. Install the required dependencies
   ```
   go get -u
   ```

## Usage

```
Usage: ./JSPyrate [OPTIONS]

Options:
  -u, --urls string        Path to the file containing the list of JavaScript URLs
  -s, --secrets string     (Optional) Path to the file containing the regex wordlist for hardcoded secrets
  -o, --output string      (Optional) Path to the output directory (default: "./output")
  -t, --threads int        (Optional) Number of threads to use for processing (default: 10)
  -h, --help               Show this help message and exit
```

## Example

```
./JSPyrate -u urls.txt -s default.txt -t 20
```

This command will analyze the JavaScript files from the URLs listed in `urls.txt`, search for hardcoded secrets using the regex patterns in `secrets.txt`, and process the files using 20 threads.

## Contributing

Contributions are welcome! If you have any ideas, improvements, or bug fixes, please submit a pull request or create an issue.

## License

JSPyrate is released under the MIT License. See [LICENSE](./LICENSE) for details.
