# Pristine-Streaming-Downloader
Pristine Streaming downloader written in Go.
![](https://i.imgur.com/Se6ZRrl.png)
[Windows, Linux, macOS, and Android binaries](https://github.com/Sorrow446/Pristine-Streaming-Downloader/releases)

# Setup
Input credentials into config file.
Configure any other options if needed.
|Option|Info|
| --- | --- |
|email|Email address.
|password|Password.
|format|Download format. 1 = 320 Kbps MP3, 2 = 16/24-bit FLAC.
|outPath|Where to download to. Path will be made if it doesn't already exist.

# Usage
Args take priority over the config file.

Download two albums:   
`ps_dl_x64.exe https://pristinestreaming.com/app/browse/albums/1885 https://pristinestreaming.com/app/browse/albums/1886`

Download a single album and from two text files in format 1:   
`ps_dl_x64.exe https://pristinestreaming.com/app/browse/albums/1885 G:\1.txt G:\2.txt -f 1`

```
 _____ _____    ____                _           _
|  _  |   __|  |    \ ___ _ _ _ ___| |___ ___ _| |___ ___
|   __|__   |  |  |  | . | | | |   | | . | .'| . | -_|  _|
|__|  |_____|  |____/|___|_____|_|_|_|___|__,|___|___|_|

Usage: ps_dl_x64.exe [--format FORMAT] [--outpath OUTPATH] URLS [URLS ...]

Positional arguments:
  URLS

Options:
  --format FORMAT, -f FORMAT
                         Download format.
                         1 = 320 Kbps MP3
                         2 = 16/24-bit FLAC. [default: -1]
  --outpath OUTPATH, -o OUTPATH
                         Where to download to. Path will be made if it doesn't already exist.
  --help, -h             display this help and exit
  ```
  
# Disclaimer
- I will not be responsible for how you use Pristine Streaming Downloader.    
- Pristine Streaming and Pristine Classical brand and names are the registered trademarks of their respective owners.    
- Pristine Streaming Downloader has no partnership, sponsorship or endorsement with Pristine Streaming or Pristine Classical.
