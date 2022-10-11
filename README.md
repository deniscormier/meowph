# Meowph photo organizer

```sh
$ meowph

NAME:
   meowph.exe - A new cli application

USAGE:
   meowph.exe [global options] command [command options] [arguments...]

COMMANDS:
   query, q   list image files that this tool can target
   rename, r  rename image files to photo-taken timestamp
   move, m    move image files into target directory
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```

## Example usage

``` sh
$ meowph query
IMG_0001.jpg
IMG_0002.jpg

$ meowph query | wc -l
2

$ meowph rename
IMG_0001.jpg -> 2022-08-14-19.30.58.jpg
IMG_0002.jpg -> 2022-08-14-19.31.23.jpg

$ meowph move -target sub_folder
2022-08-14-19.30.58.jpg -> sub_folder/2022-08-14-19.30.58.jpg
2022-08-14-19.31.23.jpg -> sub_folder/2022-08-14-19.31.23.jpg
```
