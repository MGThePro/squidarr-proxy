# squidarr-proxy

IMPORTANT: Not finished yet. Basic downloading works but many things are still broken. See the TODO section below

Complete your Lidarr library by downloading from Qobuz via squid.wtf

## setup

Build the Docker image via the included Dockerfile, then use the included docker-compose.yml as reference to create your container.

Within Lidarr, set up a new Newznab indexer with the following settings:
1. Disable RSS
2. Set the URL to the IP/Hostname of your squidarr-proxy container, but make sure it begins with http:// and ends with your configured port (8687 by default)
3. Set the API path to /indexer
4. Once your Downloader is set up, set the squidarr-proxy downloader as the downloader for this indexer

For the downloader, add a new SABnzbd downloader and configure the following:
1. Set the IP and port of the squidarr-proxy container
2. Set the Url base to "downloader"
3. Enter anything into the API Key field. This isn't actually wired up to anything yet, but Lidarr requires one to save the downloader.
Ideally set the API Key here that you set in your docker compose so your container doesn't stop working when I finally get to setting this up. Same goes for the Newznab indexer

## TODO/Things that are broken
1. Barely any metadata gets applied to the files. Pretty much just the Album name, Title, Artist and Track number. The data is available in the API with no extra requests, just need to wire it up
2. Deleting things from the History/Queue doesn't seem to quite work every time.
3. Starting a download causes it to download the file immediately, without reporting back to Lidarr. Because of this the download progress can't be tracked on Lidarr (though this worked previously)
4. Sometimes automatic importing doesn't work because Lidarr can't do the 80% match thing. Not sure what that even means.
5. Sub-optimal error handling
