# WallFetch

CLI tool to fetch wallpapers from various sources.

## APIs & Authentication
    Wallhaven:
        Public REST API at https://wallhaven.cc/api/v1/
        You’ll need to register for an API key and pass it as X-Api-Key in headers.

## Core Features
    Fetch Trending Wallpapers
        fetch wallhaven --sort toplist --categories anime --page 1

    Prevent Duplicates
        Store metadata (source + image ID + URL + local checksum) in a lightweight SQLite DB.
        On each fetch, skip any image whose source-ID or SHA256 checksum already exists.

    Manage Local Library
        list — show all downloaded wallpapers (with metadata and tags).
        prune --keep N — remove oldest until only N remain.
        browse [source] [--limit N] — preview URLs or open in browser.

    Automated Scheduling
        setup a cron job to run `wallfetch` periodically with desired options.

## Duplicate Detection
    When saving an image:
        Download bytes → compute sha256.Sum256(data) → hex string.
        SELECT 1 FROM images WHERE checksum = ? OR source_id = ?.
        If none → save file and insert metadata (source, source_id, url, checksum, downloaded_at).
