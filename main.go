package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/urfave/cli/v2"
)

// For now, I'm only interested in looking at files that have a chance of containing EXIF information
// https://en.wikipedia.org/wiki/Exif
var allowedImageFileExts = []string{".heic", ".jpg", ".jpeg", ".tif", ".tiff", ".wav", ".png", ".webp"}

type Flags struct {
	dryRun *cli.BoolFlag
	from   *cli.TimestampFlag
	to     *cli.TimestampFlag
}

type Options struct {
	dryRun bool
	from   *time.Time
	to     *time.Time
}

func main() {
	flags := Flags{
		dryRun: &cli.BoolFlag{
			Name:    "dry-run",
			Aliases: []string{"d"},
			Usage:   "log operations that would normally take place, without performing them",
		},
		from: &cli.TimestampFlag{
			Name:     "from",
			Layout:   "2006-01-02T15:04:05",
			Timezone: time.Local,
			Usage:    "targeting files after (or equal to) a photo-taken timestamp",
		},
		to: &cli.TimestampFlag{
			Name:     "to",
			Layout:   "2006-01-02T15:04:05",
			Timezone: time.Local,
			Usage:    "targeting files before (or equal to) a photo-taken timestamp",
		},
		// -folder -f (optional, default: working directory): target folder (absolute path or relative to working directory)

		// Future work: distinguish between recursive and non-recursive
		// For now, it'll be recursive and scan all sub-folders
		// &cli.BoolFlag{
		// 	Name:        "recursive",
		// 	Aliases:     []string{"r"},
		// 	Usage:       "target files in subfolders of target folders",
		// 	Destination: &options.recursive,
		// },
	}

	app := &cli.App{
		Usage: "a photo manager with a few specific capabilities",
		Commands: []*cli.Command{
			{
				Name:    "query",
				Aliases: []string{"q"},
				Usage:   "list image files that this tool can target",
				Flags: []cli.Flag{
					flags.from,
					flags.to,
				},
				Action: func(cCtx *cli.Context) error {
					options := Options{
						from: flags.from.Get(cCtx),
						to:   flags.to.Get(cCtx),
					}
					paths := cCtx.Args().Slice()
					if len(paths) == 0 {
						paths = []string{"."}
					}
					for _, path := range paths {
						path, err := filepath.Abs(path)
						ifErrFatal(err)

						err = filepath.Walk(path, handleQuery(options))
						ifErrFatal(err)
					}
					return nil
				},
			},
			{
				Name:    "rename",
				Aliases: []string{"r"},
				Usage:   "rename image files to photo-taken timestamp",
				Flags: []cli.Flag{
					flags.dryRun,
					flags.from,
					flags.to,
				},
				Action: func(cCtx *cli.Context) error {
					options := Options{
						dryRun: flags.dryRun.Get(cCtx),
						from:   flags.from.Get(cCtx),
						to:     flags.to.Get(cCtx),
					}
					paths := cCtx.Args().Slice()
					if len(paths) == 0 {
						paths = []string{"."}
					}
					for _, path := range paths {
						path, err := filepath.Abs(path)
						ifErrFatal(err)

						err = filepath.Walk(path, handleRename(options))
						ifErrFatal(err)
					}
					return nil
				},
			},
			{
				Name:    "move-wd",
				Aliases: []string{"m"},
				Usage:   "move image files into the working directory",
				Flags: []cli.Flag{
					flags.dryRun,
					flags.from,
					flags.to,
					// Future work:
					// -folder -f (optional, default: working directory): target folder (absolute path or relative to working directory)
				},
				Action: func(cCtx *cli.Context) error {
					options := Options{
						dryRun: flags.dryRun.Get(cCtx),
						from:   flags.from.Get(cCtx),
						to:     flags.to.Get(cCtx),
					}
					paths := cCtx.Args().Slice()
					if len(paths) == 0 {
						paths = []string{"."}
					}
					for _, path := range paths {
						path, err := filepath.Abs(path)
						ifErrFatal(err)

						err = filepath.Walk(path, handleMove(options))
						ifErrFatal(err)
					}
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	ifErrFatal(err)
}

func ifErrFatal(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
