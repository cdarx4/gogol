# gogol
The game of GO in Golang

# Go rules

We will base our game on those go rules :

![Go rules image](images/go-rules.webp)

This image was taken from : https://www.reddit.com/r/baduk/comments/vn936i/i_have_created_an_image_to_help_you_understand/


# TODO

[ ] Optimize the group merging
[ ] Add the KO rule detection using the zobrist hash
[ ] Add the winning condition (to be determined)
[ ] Add more depth to the OpenAI query like the difficulty, ...
[ ] Make the interface more beautiful
[ ] Add the pipeline in Github for tests, linting, ...
[ ] Publish on the github pages
[ ] add the embed for images

# Project structure

In this project, there are multiple packages that each serve a distinct purpose :

- game : Manages the game logic with the board, the stones, ...
- renderer : Manages the rendering of the game
- bot : Manages the bot that plays against the player
- main : Entry point of the application

## Game

In the game package, we've split the logic into multiple files :

- board : this file manages the board and how the stones are placed on it. This means that it handles the placement of the stones, the groups, the liberties AND handles the group deletion when a group has no liberties left.
- game : this file is a sort of controller that makes the link between the board and the renderer. It is responsible for updating the board and the renderer when a stone is placed. It doesn't handle the stones, groups, liberties, ...
- types : this files defines the types used in the game like the stones, the groups, the player, ... (the board type is defined in the board.go file)

## Renderer

The renderer package is responsible for rendering the game. It uses the ebiten library to render the game.

## Bot

The bot package is responsible for playing against the player. It uses the OpenAI API to get the next move. It sends the full board as text to the API and gets the coordinates of the next move.

# How to run the game

To run the game, you need to have go installed on your system.

Then, you can run the game by running the following command :

```bash
go run main.go
```

# How to build the game

To build the game, you can run the following command :

```bash
go build main.go
```

# Releases

This project uses [Go Releaser](https://goreleaser.com/) to automate releases. When a new version tag (e.g., `v1.0.0`) is pushed to the repository, GitHub Actions automatically builds and publishes releases for multiple platforms.

## Downloading Releases

You can download pre-built binaries from the [Releases page](https://github.com/cdarx4/gogol/releases) on GitHub. Releases are available for:

- **Linux** (amd64, arm64) - `.tar.gz` archive
- **macOS** (amd64, arm64) - `.tar.gz` archive  
- **Windows** (amd64) - `.zip` archive

Each release includes:
- Platform-specific binaries
- SHA256 checksums for verification
- README and documentation

## Installation from Release

### Linux

1. Download the appropriate `.tar.gz` file for your architecture
2. Extract the archive:
   ```bash
   tar -xzf gogol_<version>_linux_<arch>.tar.gz
   ```
3. Run the binary:
   ```bash
   ./gogol
   ```

### macOS

1. Download the appropriate `.tar.gz` file for your architecture
2. Extract the archive:
   ```bash
   tar -xzf gogol_<version>_darwin_<arch>.tar.gz
   ```
3. Run the binary:
   ```bash
   ./gogol
   ```

### Windows

1. Download the `.zip` file
2. Extract the archive
3. Run `gogol.exe`

## Creating a Release

To create a new release:

1. Update the version in your code if needed
2. Commit your changes
3. Create and push a new tag:
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
4. GitHub Actions will automatically build and publish the release

## Development

For local development and testing:

```bash
# Run the game
go run main.go

# Build locally
go build -o gogol main.go

# Run tests
go test ./...
```

