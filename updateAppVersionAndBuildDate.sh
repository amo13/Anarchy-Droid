#!/bin/bash

if [[ "$OSTYPE" == "darwin"* ]]; then
	sed -i '' "s/var AppVersion string.*/var AppVersion string = \"`awk -F'[ ="]+' '$1 == "Version" { print $2 }' FyneApp.toml`\"/" main.go
	sed -i '' "s/var BuildDate string.*/var BuildDate string = \"`date +%Y-%m-%d`\"/" main.go
else
	sed -i "s/var AppVersion string.*/var AppVersion string = \"`awk -F'[ ="]+' '$1 == "Version" { print $2 }' FyneApp.toml`\"/" main.go
	sed -i "s/var BuildDate string.*/var BuildDate string = \"`date +%Y-%m-%d`\"/" main.go
fi
