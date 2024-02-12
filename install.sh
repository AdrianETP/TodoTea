#!/bin/bash

if [ ! -d "$HOME/todotea" ]; then 
    mkdir $HOME/todotea/
    echo "[]" >> $HOME/todotea/tasks.json
fi

if [ -f "/bin/todotea" ]; then
    sudo rm /bin/todotea
fi
sudo cp ./todotea /bin/
