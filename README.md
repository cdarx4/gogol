# gogol
The game of GO in Golang

# Go rules

We will base our game on those go rules :

![Go rules image](images/go-rules.webp)

This image was taken from : https://www.reddit.com/r/baduk/comments/vn936i/i_have_created_an_image_to_help_you_understand/

# Project structure

In this project, there are multiple packages that each serve a distinct purpose :

- game : Manages the game logic with the board, the stones, ...
- renderer : Manages the rendering of the game
- main : Entry point of the application

## Game

In the game package, we've split the logic into multiple files :

- board : this file manages the board and how the stones are placed on it. This means that it handles the placement of the stones, the groups, the liberties but doesn't directly remove stones from the board if some are captured.
- game : this file manages the game logic itself. It checks the liberties, the captures, the end of the game and logs the current state of the game in the console.
- types : this files defines the types used in the game like the stones, the groups, the player, ... (the board type is defined in the board.go file)

## Renderer

The renderer package is responsible for rendering the game. It uses the ebiten library to render the game.

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

