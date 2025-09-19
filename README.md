imgdate
=======

Helper utility that renames jpeg images according to their exif data.

To use:

- Place images into the `files` directory. Sub-directories are ignored.

- Renamed images are copied into the `dest/<date>` folder. 

- The latest date is used for target folder name. 

- To append a custom title to folder name, use `-t` tag. Example:
```
go run . -t="fishing-trip"
```


