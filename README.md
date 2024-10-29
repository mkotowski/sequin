# sequin

Parse ANSI sequences into human-readable format.

```console
$ printf '\033[48;2;255;0;0m\033[m' | sequin
CSI "\x1b[48;2;255;0;0m": Background color: {255 0 0 255}
CSI "\x1b[m": Reset style
```
