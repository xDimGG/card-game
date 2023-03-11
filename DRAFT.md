# Turn-Based Game Server Architecture (Initial Draft)

## Goal
A server for turn-based card/board games

## Idea
Let server handle state/game logic and inform the clients of their legal moves and any extra info

## Pros
- Security: very difficult/impossible to cheat or do illegal moves
- Consistency: all clients should have the same state as the server at all times
- Client logic: clients will have to deal with almost no logic, as it will all be handled by the server

## Cons
- Reactivity: clients will have to wait for a server response each time they make a move

## Design
- Packet structure (symmetric)
    - Server->Client
        - Type (string) (required): message or error
        - Game (string) (required): the current game
        - State (object) (required): the state of the game
        - Moves (string[]) (required): 0 or more legal moves
        - Data (object) (optional): any extra data about the moves
    - Client->Server
        - Type (string): message
        - Moves (string[]) (required): 1 or more legal moves
        - Data (object) (optional): any extra data about the moves

## Potential Problems
- Moves that require combining two cards
