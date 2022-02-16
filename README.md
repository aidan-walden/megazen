# megazen

MegaZen is a media downloader written in Go that takes requests through a REST API. 

## Currently supported file hosts

- anonfiles.com
- bunkr.is / bunkr.to
- cyberdrop.me
- gofile.io
- putme.ga (only albums)
- pixeldrain.com

## Request endpoints

    POST /api/submit
    
    Takes a JSON array containing a JSON object for each download to be executed with the following fields:
    - url
    - password
---
    POST /api/submitBulk
    Takes a JSON array payload containing the URLs you wish to download. Assumes all submissions have no password.
---
All downloads will be output to a directory named downloads, in the same folder as the program executable.

A barebones download progress tracker can be found hosted at localhost:3000

## TODO:

- Support for mega.nz
- Config file
- Ability to exclude certain files from folder downloads
- API authorization
- React web app frontend
