# botsbox


Requirements
===

macOS:

```bash
$ sudo port install gtk3
$ sudo port install webkit2-gtk
```


Build
===

Make sure "golang.org/x/net" package has installed (Because of GFW, you need install manually in China)

```bash
$ ./build.sh
```

Run tests
===

```bash
$ go test -v ./...
```