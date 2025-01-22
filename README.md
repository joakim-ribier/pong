# PONG

The first game `Pong` was developped by Atari in 1972.

![](pong.png)

[Usage](#usage) - [How to play](#how-to-play) - [Features](#features) - [License](#license)

## Usage

Clone the project and work with the source or directly `go install`.

### Git clone

```bash
$ git clone https://github.com/joakim-ribier/pong
$ go build -o . ./...

$ ./pong
```

### Go Install

```bash
$ go install -v github.com/joakim-ribier/pong/cmd/client@latest

$ ./pong
```

### Binaries

Available only for windows from [Releases](https://github.com/joakim-ribier/pong/releases).

For the others distributions: [Is_it_possible_to_cross-compile_an_application_with_Ebitengine](https://ebitengine.org/en/documents/faq.html#Is_it_possible_to_cross-compile_an_application_with_Ebitengine?)

## How to play

### Singleplayer

```bash
# start the game with no option and enjoy it
$ ./pong
```

### Multiplayer

We should have a server which host the game and a client to play with.

The server should start before the client and It's it which handle the whole game.

#### How to start the server
```bash
$ ./pong --server 127.0.0.1:3000
```

#### How to start the client

```bash
$ ./pong --client 127.0.0.1:3000
```

## Features

### Next

...

### latest

Create a `Pong` game with [`ebitengine`](https://ebitengine.org/) 2D engine.

* [x] Implement a singleplayer mode (`Player L` VS `Player R`)
* [x] Implement a multiplayer mode (`Server` VS `Client`) with the `[UDP]` protocol to exchange messages between the server and the client
* [x] Add Github workflows to generate binaries

## License

This software is licensed under the MIT license, see [License](https://github.com/joakim-ribier/pong/blob/main/LICENSE) for more information.