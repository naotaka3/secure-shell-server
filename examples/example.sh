#!/bin/bash

# This is an example script to demonstrate the functionality of Secure Shell Server
# Only allowed commands will be executed when run through the secure shell server

# Print a message
echo "Starting script execution..."

# List files in the current directory
echo "Files in the current directory:"
ls -la

# Display the content of a file (will work if file exists and cat is allowed)
echo "Content of example.sh:"
cat example.sh

# Attempt to execute a command that might be blocked
echo "Attempting to execute 'rm' command (should be blocked):"
rm -f nonexistent_file.txt

# Print another message
echo "Script execution complete."
