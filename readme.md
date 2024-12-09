## Architecture

Troje-NG is structured into several key components:

- Container Manager: Handles container lifecycle and resource management
- Proxy: Manages connection forwarding between client and container
- Packet Capture: Monitors network traffic within containers
- Housekeeping: Automatically cleans up idle containers

## Security Considerations

- Containers are resource-limited and security-constrained
- Each attacker gets an isolated environment
- Network traffic is monitored and logged
- Containers are automatically destroyed after inactivity
- Base container uses minimal Alpine Linux installation

## Contributing

Contributions are welcome! Please submit pull requests to the [GitHub repository](https://github.com/referefref/troje-ng).

## Credits

**Original Author**
- Remco Verhoef ([@remco_verhoef](https://twitter.com/remco_verhoef))

**Current Maintainer**
- referefref

## Changes from Original

- Complete rewrite in modern Go
- Structured project layout
- Proper error handling and logging
- Container resource management
- Network traffic capture
- Automatic container cleanup
- Alpine Linux base instead of Ubuntu
- Improved security constraints
- Better configuration options

## License

Code and documentation copyright 2024 referefref.
Original code copyright 2011-2014 Remco Verhoef.
Code released under [the MIT license](LICENSE).

## Warning

This software is provided for research purposes only. Ensure proper isolation and monitoring when deploying in any environment.
