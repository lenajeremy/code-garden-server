#!/bin/sh

# Check if a language argument is provided
if [ -z "$1" ]; then
    echo "Error: No language specified."
    echo "Usage: $0 <language>"
    exit 1
fi

LANGUAGE="$1"

# Determine the file extension and command based on the language
case "$LANGUAGE" in
    go)
        FILE="file.go"
        RUN_CMD="go run $FILE"
        ;;
    python)
        FILE="file.py"
        RUN_CMD="python3 $FILE"
        ;;
    rust)
        FILE="file.rs"
        RUN_CMD="rustc $FILE && ./file"
        ;;
    cpp)
        FILE="file.cpp"
        RUN_CMD="g++ $FILE && ./a.out"
        ;;
    swift)
        FILE="file.swift"
        RUN_CMD="swift $FILE"
        ;;
    ruby)
        FILE="file.rb"
        RUN_CMD="ruby $FILE"
        ;;
    node)
        FILE="file.js"
        RUN_CMD="node $FILE"
        ;;
    *)
        echo "Error: Unsupported language '$LANGUAGE'."
        exit 1
        ;;
esac

# Read the input source code and write to the appropriate file
cat > "$FILE"

# Run the code using the specified command
eval "$RUN_CMD"

# Delete the source code file after execution
rm "$FILE"
