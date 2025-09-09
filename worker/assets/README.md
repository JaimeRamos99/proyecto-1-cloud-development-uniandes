# Assets Directory

## Watermark (watermark.png)

This directory should contain the ANB watermark image used by the video processor.

### Requirements:

- **File name**: `watermark.png`
- **Format**: PNG with transparency
- **Recommended size**: 200x50 pixels
- **Content**: ANB logo or text

### Options to create the watermark:

#### Option 1: Use the provided script (requires ImageMagick)

```bash
# Install ImageMagick in the Docker container
# Add to Dockerfile: RUN apk add --no-cache imagemagick

# Run the script
./create_watermark.sh
```

#### Option 2: Create manually

Create a PNG file with:

- White text "ANB"
- Semi-transparent background
- Black outline for visibility
- Size: 200x50 pixels

#### Option 3: Use existing ANB logo

Replace `watermark.png` with your official ANB logo in PNG format.

### Placement

The watermark will be placed in the **top-right corner** of all processed videos, with a 10-pixel margin from the edges.

### If watermark is missing

The video processor will continue processing videos without the watermark and log a warning message.
