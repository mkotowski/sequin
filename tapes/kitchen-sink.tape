Output kitchen-sink.gif
Set Width 1400
Set Height 825

# Type "\x1b_some_command\x1b\\\x1bP=1s\x1b\\\x1b]0;Sequin!\a\x1b[?25l\x1b[?2004h\x1b[2J\x1b7\x1b[H\rWho’s hungry\x1b\x1b[K\r\n\x1b[K\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b8\x1b\x1b[999;99;99\x1b"

Sleep 2s
Type "printf '"
Sleep 500ms
Hide
Type "\x1b]0;Sequin!\a\x1b[?25l\x1b[?2004h\x1b[2J\x1b7\x1b[H\rWho’s hungry?\x1b\x1b[K\r\n\x1b[K\x1b[D\x1b[2K\r\x1b[?2004l\x1b[?25h\x1b8\x1b"
Show
Sleep 1s
Type "'"
Sleep 250ms
Type " | sequin"
Sleep 150ms
Enter
Sleep 22s
