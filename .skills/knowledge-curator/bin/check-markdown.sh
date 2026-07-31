#!/bin/bash

# Markdown format check script
# Auto-check format compliance for knowledge base content

echo "Markdown format check starting..."

# Check file path
FILE_PATH="$1"
if [ -z "$FILE_PATH" ]; then
    echo "ERROR: file path is required"
    echo "Usage: ./check-markdown.sh <file_path>"
    exit 1
fi

if [ ! -f "$FILE_PATH" ]; then
    echo "ERROR: file not found: $FILE_PATH"
    exit 1
fi

echo "Checking file: $FILE_PATH"

# Check ordered list format (newline required after each item number)
echo "  [1/3] Checking list format..."
if grep -E "^[0-9]+\. [^\n]*[0-9]+\." "$FILE_PATH"; then
    echo "FAIL: ordered list items must be separated by newlines"
    echo "  Each numbered item should be on its own line to avoid rendering on the same line"
    exit 1
fi

# Check code block language tags
echo "  [2/3] Checking code block language tags..."
CODE_BLOCKS=$(grep -c "^```" "$FILE_PATH" || true)
LANGUAGE_BLOCKS=$(grep -c "^```[a-z]" "$FILE_PATH" || true)

if [ "$CODE_BLOCKS" -gt 0 ] && [ "$LANGUAGE_BLOCKS" -lt "$CODE_BLOCKS" ]; then
    echo "WARN: some code blocks are missing language tags"
    echo "  Consider adding language identifiers to all code blocks"
fi

# Check paragraph length (rough check)
echo "  [3/3] Checking paragraph length..."
LONG_PARAGRAPHS=$(awk 'BEGIN{RS=""; count=0} length($0) > 300 {count++} END{print count}' "$FILE_PATH")
if [ "$LONG_PARAGRAPHS" -gt 0 ]; then
    echo "WARN: long paragraphs detected (over 300 chars)"
    echo "  Consider splitting long paragraphs into 2-4 sentence chunks"
fi

echo "Markdown format check completed."
echo "  See: .skills/blog-writer/references/writing-style-guide.md"
echo "  See: .skills/knowledge-curator/references/markdown-checklist.md"