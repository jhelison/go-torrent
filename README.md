<a name="readme-top"></a>
![GitHub License](https://img.shields.io/github/license/jhelison/go-torrent)
![GitHub release (with filter)](https://img.shields.io/github/v/release/jhelison/go-torrent)

<br />
<div align="center">
  <h3 align="center">Go Torrent</h3>

  <p align="center">
    A small and optimized go torrent downloader
    <br />
  </p>
</div>



<!-- TABLE OF CONTENTS -->
- [About The Project](#about-the-project)
  * [Built With](#built-with)
- [Getting Started](#getting-started)
  * [Installation](#installation)
    + [Downloading Binaries from GitHub Releases](#downloading-binaries-from-github-releases)
    + [Build from source](#build-from-source)
- [Usage](#usage)
  * [Configuration](#configuration)
  * [Basic Commands](#basic-commands)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)
- [Acknowledgments](#acknowledgments)



<!-- ABOUT THE PROJECT -->
## About The Project

This is a Go implementation based on the guide from [Jessi Li](https://blog.jse.li/posts/torrent/) with a few improvements.

Here's what we encompass in this project:
- Bencode parser 📃
- P2p torrent download using Bittorrent protocols 🔒
- Cobra CLI and Viper configuration 🔧

From Jessi Li orinals implementation we have the following:
- Chunk based file writes
- Viper and Cobra 🤖
- Improvements on memory management 🏎️
- Better pieces handling  

<p align="right">(<a href="#readme-top">back to top</a>)</p>


### Built With

One of the most awesome things about this project, is that it's was built using GO!

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- GETTING STARTED -->
## Getting Started

No pre-requisites are necessary to run Go Torrent.

### Installation

You have two options to get Go Torrent running:
- Downloading the binaries
- Build from source

#### Downloading Binaries from GitHub Releases

1. Visit the [Releases page](https://github.com/jhelison/go-torrent/releases/latest) of the go-torrent GitHub repository.

2. Choose the latest release.

3. Under "Assets", download the appropriate binary for your operating system (Linux, Mac, Windows).

4. After downloading, you may need to grant execution permissions to the binary. For Linux or Mac, use:

```bash
chmod +x go-torrent
```

5. Place the binary in a directory within your PATH for easy access.

#### Build from source

Prerequisites:
- Ensure you have Go installed on your system. You can download it from the [official Go website](https://go.dev/doc/install).


To install go-torrent from source, follow these steps:

1. Clone the repository from GitHub:

```bash
git clone https://github.com/jhelison/go-torrent.git
```

2. Navigate to the cloned directory:

```bash
cd go-torrent
```

3. Build the application:

```bash
make build

or

mkdir build
go build -o build
```

4. Place the binary in a directory within your PATH for easy access.

<!-- USAGE EXAMPLES -->
## Usage

### Configuration

- `go-torrent` uses a configuration file located at $HOME/.go-torrent.toml. If this file does not exist, the program will create a default one on first run.
- You can modify this file to change settings like download path, log level, and peer settings.

### Basic Commands

To start go-torrent, simply run the binary:

```bash
go-torrent
```

To download a torrent file, use the `download` command followed by the path to the torrent file:

```bash
go-torrent download /path/to/torrentfile.torrent
```

You can specify the output directory for downloads using the `--output` flag:

```bash
go-torrent download /path/to/torrentfile.torrent --output /path/to/download/directory
```

**Global flags**

- Specify a custom configuration file:

```bash
go-torrent --config /path/to/config.toml
```

- Set the log level (trace|info|warn|err|disabled):

```bash
go-torrent --log-level info
```

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ROADMAP -->
## Roadmap

Still lot's to be done!

- [ ] Add magnetic link support
- [ ] Add multi-torrent download
- [ ] Improve peer management

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTRIBUTING -->
## Contributing

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

Don't forget to use conventional comments and write a good Pull Request 😊

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE.txt` for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- CONTACT -->
## Contact

Jhelison Uchoa - [Linkedin](https://www.linkedin.com/in/jhelison/) - jhelisong@gmail.com    

<p align="right">(<a href="#readme-top">back to top</a>)</p>



<!-- ACKNOWLEDGMENTS -->
## Acknowledgments

Use this space to list resources you find helpful and would like to give credit to.

* [Building a BitTorrent client from the ground up in Go](https://blog.jse.li/posts/torrent/)

<p align="right">(<a href="#readme-top">back to top</a>)</p>
