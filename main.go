package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

// For now, I'm only interested in looking at files that have a chance of containing EXIF information
// https://en.wikipedia.org/wiki/Exif
var allowedImageFileExts = []string{".heic", ".jpg", ".jpeg", ".tif", ".tiff", ".wav", ".png", ".webp"}

type Flags struct {
	dryRun *cli.BoolFlag
	from   *cli.TimestampFlag
	to     *cli.TimestampFlag
	target *cli.PathFlag
}

type Options struct {
	dryRun bool
	from   *time.Time
	to     *time.Time
	target string
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
		target: &cli.PathFlag{
			Name:      "target",
			Value:     ".",
			TakesFile: true,
			Usage:     "target folder, an absolute path or relative to working directory (default: working directory)",
		},
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
					cmdPaths := cCtx.Args().Slice()
					options := Options{
						from: flags.from.Get(cCtx),
						to:   flags.to.Get(cCtx),
					}
					return handleQuery(cmdPaths, options)
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
					cmdPaths := cCtx.Args().Slice()
					return handleRename(cmdPaths, options)
				},
			},
			{
				Name:    "move",
				Aliases: []string{"m"},
				Usage:   "move image files into target directory",
				Flags: []cli.Flag{
					flags.dryRun,
					flags.from,
					flags.to,
					flags.target,
				},
				Action: func(cCtx *cli.Context) error {
					options := Options{
						dryRun: flags.dryRun.Get(cCtx),
						from:   flags.from.Get(cCtx),
						to:     flags.to.Get(cCtx),
					}
					cmdTargetPath := flags.target.Get(cCtx)
					var err error
					options.target, err = filepath.Abs(cmdTargetPath)
					if err != nil {
						return errors.Wrapf(err, "failure converting target path %s to absolute path", cmdTargetPath)
					}
					cmdPaths := cCtx.Args().Slice()
					return handleMove(cmdPaths, options)
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
