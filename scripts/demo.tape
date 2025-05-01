Output docs/demo.gif

Set Shell "bash"
Set FontSize 14
Set Width 1200
Set Height 800
Set Framerate 24
Set Padding 5

Type `go run ./app/main.go` Enter
Sleep 2s

# Show help
Type "?" Sleep 0.5s
Screenshot docs/help.png Sleep 1s

# Hide help
Type "?" Sleep 0.5s

# Navigate to search
Type "ctrl+f" Sleep 0.5s
Screenshot docs/search.png Sleep 1s

# Press ESC to go back
Type "esc" Sleep 0.5s

# Show filter
Type "/" Sleep 0.5s
Screenshot docs/filter.png Sleep 1s

# Close filter
Type "esc" Sleep 0.5s

# Show debug info (memory stats)
Type "ctrl+d" Sleep 1s
Screenshot docs/memory-stats.png Sleep 1s

# Trigger quit prompt
Type "ctrl+c" Sleep 0.5s
Screenshot docs/quit.png Sleep 1s

# Cancel quit
Type "n" Sleep 0.5s

# Quit the application
Type "ctrl+c" Sleep 0.5s
Type "y" Sleep 0.5s
