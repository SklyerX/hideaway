# Hideaway

### What is Hideaway?

Hideaway is a CLI tool that encrypts any file located on your computer and stores it securely until you retrieve it. All files and information are stored locally and the app **never** contacts anything outside of your local environment. Fully local, fully offline.

### Why Did I build this project?

Recently I've been getting more interested in Go and personally I learn better when building something that I like rather than traditional courses / crash courses. I take those to learn the basics but for me I learn through building projects over and over and over again no matter how useless or useful it maybe. I must admit at times it was challenging to continue with the project even tho it took a lot of Googling and documentation and having Claude explain the Go logic and everything to me, in the end I feel better knowing I took more steps towards learning Go.

So flame me if you must if this is not conventional Go code, if it feels to beginner, or not to your liking but to me this project was a step forward :)

---

Now the good stuff.


## Installation

Because Go is (Go)ated all you have to do is simply run:
```
go install github.com/sklyerx/hideaway/@latest
```

and your done!

## Usage

The base command to launch Hideaway is just:
```
hideaway
```

### Init

When you first install Hideaway it needs to initiate some default settings / locations, to do this run:

```
hideaway init
```


this will prompt you for a master password, this is **not** stored in plain text, **do not share this with anyone**, and there is no *reset password* option so make sure you have it committed to memory.

### Starting

Run the base command shown above, enter your password, that will start an interactive repl. Here you get access to the SUDO commands that are not available before initializing or authentication.

Run the `help` command for more information but here is a quick breakdown:

```
❯ hideaway
Enter password: 
Welcome to Hideaway Repl!
Type 'help' for available commands or 'exit' to quit.
```

```
add <filePath> --delete OR -d
list
stats
```

### Resetting

In the worst case, if you have forget your master-password you can run:

```
hideaway reset
```

This will completely wipe any data stored on your device from Hideaway this includes:
- Files
- Configurations
- Password hash
- Metadata

There is no way to reverse this, once you run the command and accept the confirmation message all data will be lost forever.

I will write a blog explaining the logic more on how the app works and why I decided to build it, either checkout [skylerx.ir](https://skylerx.ir/blog) for when it drops or check back here for more updates.