#!/bin/bash
# Extract emojis to image files using ImageMagick

# Directory to save extracted emojis
OUTPUT_DIR="share/emoji"

# Emoji font and size settings
EMOJI_FONT="Noto Color Emoji"
EMOJI_SIZE=30

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Array of emojis to extract
EMOJIS=(
    "✨"
    "☁️"
    "🌫"
    "🌧"
    "🌧"
    "❄️"
    "❄️"
    "🌦"
    "🌦"
    "🌧"
    "🌧"
    "🌨"
    "🌨"
    "⛅️"
    "☀️"
    "🌩"
    "⛈"
    "⛈"
    "☁️"
    "🌑"
    "🌒"
    "🌓"
    "🌔"
    "🌕"
    "🌖"
    "🌗"
    "🌘"
)

# Function to extract emojis
extract_emojis() {
    for emoji in "${EMOJIS[@]}"; do
        output_file="$OUTPUT_DIR/$emoji.png"
        echo "Extracting $emoji to $output_file"
        
        convert \
            -background black \
            -size "${EMOJI_SIZE}x${EMOJI_SIZE}" \
            -set colorspace sRGB \
            "pango:<span font='${EMOJI_FONT}' size='20000'>$emoji</span>" \
            "$output_file"
            
        if [ $? -eq 0 ]; then
            echo "Successfully extracted $emoji"
        else
            echo "Failed to extract $emoji"
        fi
    done
}

# Execute the extraction
extract_emojis
