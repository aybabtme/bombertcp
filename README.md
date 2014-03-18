# bombertcp

A TCP [bomberman](https://github.com/aybabtme/bomberman) player. The player must
be compiled with the bomberman game, which will then serve as a server for TCP
clients.

Clients to this TCP player can control a bomberman player by sending various
commands on the TCP connection.  They will also receive every update to the game
from that same connection.

This is the simplest way to implement a [bomberman](https://github.com/aybabtme/bomberman)
player in your programming language of choice.

# Clients

* `bombermanpy` by uiri: https://github.com/uiri/bombermanpy
* `your client here`

# Docs

## Sending a move

You can send one move per turn to the game.

### Format

There are only 5 valid moves a client can send to the TCP server:

* `up\n`: ask to move your player up.
* `down\n`: ask to move the player down.
* `left\n`: ask to move your player left.
* `right\n`: ask to move your player right.
* `bomb\n`: ask to place a bomb at your player's current location.

Each string needs to be followed by a `\n` character, otherwise your move will
be considered invalid.

### Invalid moves

If the move you send is invalid, it will simply be ignored by the game. Only one
move per turn will be used by the game. If the string you send to the game is
not a valid string, it will be ignored.

### Buffering

If you send more than one move per turn, some of the move you send will be
buffered and sent at each subsequent turns. Other moves that aren't buffered
will be dropped on the floor and never be looked at, ever.

The length of the move buffer is specified as constant `BufferedMoves` in
`player.go`.

## Receiving an updated game state

Everytime a turn changes in the game, you will receive a JSON string that
represents all of the new game state.  This game state contains information
about your player and the current status of the board. You do not receive details
about other players, aside from their position on the board.

### Format

```JSON
{
  "Turn": 12,
  "TurnDuration": 33000000,
  "Name": "p2",
  "X": 49,
  "Y": 21,
  "LastX": 48,
  "LastY": 21,
  "Bombs": 0,
  "MaxBomb": 3,
  "MaxRadius": 3,
  "Alive": true,
  "GameObject": {"...", "..."},
  "Message": "playing",
  "Board": [
    [ "... array of Cell objects ..." ],
    [ {"Name":"Wall"}, {"Name":"p1"}, {"Name":"Ground"}, {"Name":"Rock"}, "..." ],
    "...",
    [ "... array of Cell objects ..." ]
  ]
}
```
Of course, all the `"..."` are not actually there.

* `Turn`: current turn number.
* `TurnDuration`: in nanoseconds, duration of a turn.
* `Name`: your name.
* `X`: your X coordinate on the board.
* `Y`: ...
* `LastX`: the X coordinate you had last turn.
* `LastY`: ...
* `Bombs`: current bombs you have used.
* `MaxBomb`: maximum bombs you can use.
* `MaxRadius`: explosion radius of bombs you will place in the future.
* `Alive`: whether your player is alive or dead.
* `GameObject`: unspecified, do not rely on the data of this field.
* `Message`: messages from the game, like victory, draw, etc.
* `Board`: the visible state of the board.

Each update will be a JSON object such as above, on a single line, followed by
two newline characters, as in `{json object here}\n\n`.

### The `Board`

The size of the board is not explicitely given to you, rather implicit in the
size of the 2x2 `Board` array. The size of the board can change from a game to
another, so do no hardcode them.

The position of a `Cell` is implied from it's
position in the `Board` field.  The `X` and `Y` coordinates are 0-indexed, with
`0, 0` being the top-left corner.

```
  (0, 0)  ....... (maxX, 0)
    :                  :
    :                  :
    :                  :
    :                  :
    :                  :
(0, maxY) ....... (maxX, maxY)
```

### Buffering

If you are too slow to read the updates from the TCP connection, the following
updates will be dropped and the most recent ones sent.  Some updates will be
buffered to give you a chance to read them.  The size of that buffer is
specified in `player.go`, as the constant `BufferedMoves`.

## Data dump

Some data dumped from `netcat` listening on the TCP server is available in the
`dump.json` file.
