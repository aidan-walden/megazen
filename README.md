# megazen

MegaZen is a media downloader written in Go that takes requests through a REST API. 

## Currently supported file hosts

- bunkr.is / bunkr.to
- cyberdrop.me
- gofile.io
- putme.ga (only albums)

## Request endpoints

    POST /api/submit
Takes a JSON array payload containing the URLs you wish to download. You may mix and match hosts. All downloads will be output to a directory named downloads, in the same folder as the program executable.
## TODO:

- Support for mega.nz
- React web app frontend