# goxtxt
A xmpp twtxt bot written in go

## about

This bot enables you to tweet with [twtxt][2] by sending a message via [xmpp][3]. 
Created with try and error hacking on [this example][1].
This is the first programming I did since programming some µC in 
C and programming PLCs during my studies and I have to admit I didn't take 
the time to properly dive into programming.

So consider this a study to see how far you can get with some very basic 
knowledge, 'trial and error' and searching in the docs. Anyway 
recommendations how to do better are welcome.

## requirements

* [go][4]
* [twtxt client][5]
* [coreutils][6]
* [util-linux][7]

## installation

If you have *[GOPATH][8]* set just run this command:

```
$ go install github.com/mdosch/goxtxt
```

You will find the binary in `$GOPATH/bin` or, if set, `$GOBIN`.

## configuration

The configuration is expected at `$HOME/.config/goxtxt/config.json` with this format:

```
{
    "Address": "example.com:5222",
    "BotJid": "bot@example.com",
    "Password": "ChangeThis!",
    "ControlJid": "user@example.com",
    "Twtxtnick": "mdosch",
    "TimelineEntries": 10,
    "MaxCharacters": 140
}
```

[1]:https://github.com/processone/gox/blob/master/cmd/xmpp_echo/xmpp_echo.go
[2]:https://github.com/buckket/twtxt/
[3]:https://xmpp.org/
[4]:https://golang.org/
[5]:https://github.com/buckket/twtxt
[6]:http://www.gnu.org/software/coreutils/coreutils.html
[7]:https://git.kernel.org/pub/scm/utils/util-linux/util-linux.git/about/
[8]:https://github.com/golang/go/wiki/SettingGOPATH
