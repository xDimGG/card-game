# Card Game App
I wanted to try creating my own turn-based game protocol and implement a couple of games. This was the result! The app is currently live and can be accessed
[here](https://cg.dim.codes/). Feel free to have fun with your friends or some bots!

## Design Choices
### [Go](https://go.dev/)
Go is just a personal favorite of mine. Although Go is quite basic compared to other languages with some limited standard library capabilities and error handling,
it has just enough of the things I need to create a fast and concurrent WebSocket backend. I've been using Go for a few years now, and writing in
Go doesn't get old.

### [Vue](https://vuejs.org/)
I've always been a fan of Vue despite not using it myself. Having contributed to some Vue and React projects in the past, I was able to pick up and use
Vue pretty quickly. At first, this app was built with a single HTML file and without Node, but after implementing just two games and seeing
the file quickly grow to 400+ lines, I decided to build the app with [Vite](https://vitejs.dev/). I was pleasantly surprised with how simple the conversion
process was and continued to expand my site this way. The old single-page HTML 

## Running the App
If you have Docker, running the app can be done with just `docker-compose up` in the `Server` folder.

The app comes with a pre-built frontend in the `Server` folder. If you would like to rebuild the frontend yourself, run the following commands in the
`Website` folder.
```
npm install     # installs all the project dependencies
npm run export  # runs vite export and moves the compiled files to the Server folder
```

## TODO
- Make AI for `The Mind` and `Halli Galli` better.
- Consider adding a text chat.
- Implement more games.
