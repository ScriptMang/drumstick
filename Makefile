.PHONY: install-homebrew install-dependencies import-drumstick

#TODO Add windows setup

macos-setup: install-homebrew install-dependencies import-drumstick

install-homebrew:
	/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

install-dependencies:
	brew install go@1.24.2, postgresql@16.8

# create drumstick table and import data from sql file 
import-drumstick:
	createdb drumstick
	psql --username=user drumstick < drumstick.sql
