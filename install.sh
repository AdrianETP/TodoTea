#!/bin/bash

if [ ! -d "$HOME/todotea" ]; then 
    mkdir $HOME/todotea/
    echo "[]" >> $HOME/todotea/tasks.json
fi

if [ -f "/bin/todotea" ]; then
    sudo rm /usr/local/bin/todotea
fi
sudo cp ./todotea /usr/local/bin
