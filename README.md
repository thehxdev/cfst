# CFST
This project is a fork of [CloudflareSpeedTest](https://github.com/XIU2/CloudflareSpeedTest).
You can find Cloudflare IP list (Both v4 and v6) on [ips](https://www.cloudflare.com/ips) page of Cloudflare website.

### What is different from original version?
- Translated to English
- Removed all Chinese comments
- Write tested (ping test) IPs to a file to prevent multiple ping checks. Just use already tested IPs!


## Build
To build `cfst` you need go compiler and `make` command availabe.
```bash
make
```

The command above will build the project and produce `cfst` executable.


## Usage
To get a help message:
```bash
./cfst -h
```


## Credits
Thanks to developers and all contributors of [XIU2/CloudflareSpeedTest](https://github.com/XIU2/CloudflareSpeedTest) project.


## License
The GPL-3.0 License.
