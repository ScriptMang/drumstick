.PHONY: install-homebrew install-dependencies import-drumstick

#TODO Add windows setup

macos-setup: install-homebrew install-dependencies import-drumstick

install-homebrew:
	/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

install-dependencies:
	brew install go@1.25.0, postgresql@16.10

# create drumstick table and import data from sql file 
import-drumstick:
	createdb drumstick
	psql --username=user drumstick < drumstick.sql

build:
	go build -o drumstick main.go
