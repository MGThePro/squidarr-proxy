# squidarr-proxy

Complete your Lidarr library by downloading from Qobuz via squid.wtf

## Setup

Use the included [docker-compose](docker-compose.yml) as reference for creating your container.

Within Lidarr, set up a new Newznab indexer with the following settings:
1. Disable RSS
2. Set the URL to the IP/Hostname of your squidarr-proxy container, but make sure it begins with http:// and ends with your configured port (8687 by default)
3. Set the API path to /indexer
4. Set the API token you set in your docker-compose.yml

For the downloader, add a new SABnzbd downloader and configure the following:
1. Set the IP and port of the squidarr-proxy container
2. Set the Url base to "downloader"
3. Configure the API token you set in your docker-compose.yml
4. Set this downloader as the default for the squidarr-proxy indexer
