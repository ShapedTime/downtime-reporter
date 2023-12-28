# Downtime Reporter

Downtime Reporter is a Go application that queries Prometheus metrics and generates CSV reports based on the returned data. This tool is useful for monitoring and analyzing service availability and downtime.

## Features

- Query Prometheus for historical data.
- Generate concise downtime reports in CSV format.
- Customizable time range, step intervals, and output file naming.
- Options for concise output and Unix time formatting.

## Prerequisites

- Go environment.
- Access to a Prometheus server.

## Installation

Clone the repository and build the application:

```bash
git clone [repository URL]
cd [repository directory]
go build -o dt-reporter
```

# Usage
Run the application with the necessary flags. Here is an example command:
`./dt-reporter -p <prometheus_url> -q "<query>" -c -f output.csv -st "YYYY-MM-DD HH:MM:SS" -et "YYYY-MM-DD HH:MM:SS"`

## Flags
- -p, --prometheus: Prometheus server URL (default "http://localhost:9090").
- -s, --step: Step interval in seconds/minutes (default 2m).
- -u, --unix: Show in Unix time (default false).
- -q, --query: Prometheus query, should return 0 or 1 for downtime (required).
- -st, --startTime: Start time for the query in "YYYY-MM-DD HH:MM:SS" format.
- -et, --endTime: End time for the query in "YYYY-MM-DD HH:MM:SS" format.
- -c, --concise: Concise output, only show downtime periods (default false).
- -f, --fileName: Output file name (default "output.csv").
  
## Example
`./dt-reporter -p http://localhost:9090 -q 'up{job="your_job"}' -c -f downtime.csv -st "2023-01-01 00:00:00" -et "2023-01-02 00:00:00"`

This command will query the Prometheus server for the specified job's availability, producing a concise downtime report named downtime.csv for the specified 24-hour period.

# Contributing
Contributions are welcome! Please fork the repository and submit a pull request with your changes or improvements.

