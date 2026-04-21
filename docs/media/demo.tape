# Regenerate docs/media/demo.gif via VHS.
#
# Requires: vhs, ttyd, ffmpeg.
#   go install github.com/charmbracelet/vhs@latest
#   sudo apt install ttyd ffmpeg   # macOS: brew install ttyd ffmpeg
#
# Then:
#   go build -o ./dockpose ./cmd/dockpose
#   vhs docs/media/demo.tape
#
# This script drives the real binary in --demo mode. The
# `dockpose-record` Go tool is the alternative path used in CI when ttyd
# is unavailable; it produces the same .cast file directly.

Output docs/media/demo.gif

Set FontSize 14
Set Width 1200
Set Height 740
Set Theme "GitHub Dark"
Set Padding 20
Set TypingSpeed 80ms

Type "./dockpose --demo"
Enter
Sleep 1500ms

# Walk down the stack list.
Down
Sleep 600ms
Down
Sleep 600ms
Down
Sleep 1200ms

# Open monitoring detail.
Enter
Sleep 1800ms
Down
Sleep 600ms
Down
Sleep 1000ms

# Inspect a container.
Type "i"
Sleep 1800ms
Escape
Sleep 600ms

# Back to list, open help.
Escape
Sleep 600ms
Type "?"
Sleep 1800ms
Escape

# Filter for "media".
Type "/"
Type "media"
Sleep 800ms
Enter
Sleep 800ms
Escape

# Try a destructive action -> demo toast.
Type "d"
Sleep 1500ms

# Quit.
Type "q"
