# squidarr-proxy

Complete your Lidarr library by downloading from Qobuz via squid.wtf
Almost ready for use. Try it out, tell me what's wrong. But careful - here be dragons :)

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
1. API Tokens aren't implemented yet, making every instance completely open
2. Sub-optimal error handling. At least delete incomplete downloads on errors and clear the incomplete folder in startup
3. Once it's ready for prime time - Create the container image via GitHub and make this easy to use
