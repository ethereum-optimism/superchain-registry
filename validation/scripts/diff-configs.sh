#!/bin/bash

# Check if two directories are provided
if [ $# -ne 2 ]; then
  echo "Usage: $0 <directory1> <directory2>"
  exit 1
fi

# Assign input directories to variables
dir1=$1
dir2=$2

# Check if both directories exist
if [ ! -d "$dir1" ] || [ ! -d "$dir2" ]; then
  echo "Both arguments must be directories"
  exit 1
fi

# Initialize a variable to buffer the output
output=""
overall_failure=false

# Iterate over files in the first directory
set +e # disable exit on error
for file1 in "$dir1"/*; do
  filename=$(basename "$file1")
  file2="$dir2/$filename"

  # Check if the file exists in the second directory
  if [ -e "$file2" ]; then
    output+="Comparing $file1 and $file2\n"
    diff_output=$(diff -u "$file1" "$file2")
    diff_status=$?
    if [ $diff_status -eq 1 ]; then
      overall_failure=true
      output+="Files $file1 and $file2 are different:\n$diff_output\n"
    elif [ $diff_status -eq 2 ]; then
      output+="An error occurred while comparing $file1 and $file2\n"
    fi
  else
    output+="File $filename does not exist in $dir2\n"
  fi
done

# Print the buffered output
echo -e "$output"

if [ "$overall_failure" = true ]; then
  echo "Exiting with status 1 because at least one difference was found"
  exit 1
fi
