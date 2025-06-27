# squidarr-proxy

Complete your Lidarr library by downloading from Qobuz via squid.wtf

## Setup

Build the Docker image via the included Dockerfile, then use the included docker-compose.yml as reference to create your container.

Within Lidarr, set up a new Newznab indexer with the following settings:
1. Disable RSS
2. Set the URL to the IP/Hostname of your squidarr-proxy container, but make sure it begins with http:// and ends with your configured port (8687 by default)
3. Set the API path to /indexer
4. Set the API token you set in your docker-compose.yml
4. Once your Downloader is set up, set the squidarr-proxy downloader as the downloader for this indexer

For the downloader, add a new SABnzbd downloader and configure the following:
1. Set the IP and port of the squidarr-proxy container
2. Set the Url base to "downloader"
3. Set the API token you set in your docker-compose.yml

## TODO
1. Cancelling downloads or downloads failing could still be a little problematic, though neither are very likely (I think?)
2. Once it's ready for prime time - Create the container image via GitHub and make this easy to use
