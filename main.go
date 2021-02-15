package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

var (
	reader *bufio.Scanner
	writer *bufio.Writer
)

type Cutter interface {
	cut() error
}

type params struct {
	delimiter  string
	fields     []int
	whitespace bool
	suppress   bool
	parallel   int
}

func NewCutter(
	delimiter string,
	fields string,
	whitespace bool,
	suppress bool,
	parallel int,
) (*params, error) {
	fieldsSlice, err := parseFields(fields)
	if err != nil {
		return nil, err
	}
	c := &params{
		delimiter:  delimiter,
		fields:     fieldsSlice,
		whitespace: whitespace,
		suppress:   suppress,
		parallel:   parallel,
	}
	return c, nil
}

func (c *params) cut() error {
	if c.parallel > 0 {
		return c.extractParallel()
	}
	return c.extract()
}

func (c *params) extract() error {
	lastField := c.fields[len(c.fields)-1]
	var tokens []string

	for reader.Scan() {
		line := reader.Text()

		switch {
		case !strings.Contains(line, c.delimiter):
			if c.suppress {
				continue
			}
			fmt.Fprint(writer, line, "\n")

		default:
			if c.whitespace {
				tokens = strings.Fields(line)
			} else {
				tokens = strings.Split(line, c.delimiter)
			}

			if lastField >= len(tokens) {
				return fmt.Errorf("field %d out of range in line containing %d fields", lastField+1, len(tokens))
			}

			for _, field := range c.fields[:len(c.fields)-1] {
				fmt.Fprint(writer, tokens[field], c.delimiter)
			}
			fmt.Fprint(writer, tokens[c.fields[len(c.fields)-1]], "\n")
		}
	}

	if err := reader.Err(); err != nil {
		return err
	}
	return nil
}

func (c *params) extractParallel() error {
	lastField := c.fields[len(c.fields)-1]

	in := make(chan string)
	out := make(chan string)
	semaphore := make(chan struct{}, c.parallel)
	done := make(chan struct{})
	var g errgroup.Group

	g.Go(func() error {
		defer close(in)
		for reader.Scan() {
			in <- reader.Text()
		}
		return reader.Err()
	})

	g.Go(func() error {
		for line := range in {
			line := line

			g.Go(func() error {
				semaphore <- struct{}{}

				switch {
				case !strings.Contains(line, c.delimiter):
					if c.suppress {
						return nil
					}
					out <- line

				default:
					var tokens []string
					if c.whitespace {
						tokens = strings.Fields(line)
					} else {
						tokens = strings.Split(line, c.delimiter)
					}

					if lastField >= len(tokens) {
						return fmt.Errorf("field %d out of range in line containing %d fields", lastField+1, len(tokens))
					}

					subset := make([]string, len(c.fields))
					for i, f := range c.fields {
						subset[i] = tokens[f]
					}
					joined := strings.Join(subset, c.delimiter)

					out <- joined
				}

				<-semaphore
				return nil
			})
		}
		return nil
	})

	go func() {
		for joined := range out {
			fmt.Fprint(writer, joined, "\n")
		}
		done <- struct{}{}
	}()

	err := g.Wait()
	if err != nil {
		return err
	}
	close(out)
	<-done

	return nil
}

func parseFields(s string) ([]int, error) {
	var fields []int
	for _, f := range strings.Split(s, ",") {
		if strings.Contains(f, "-") {
			r := strings.Split(f, "-")
			lo, err := strconv.Atoi(r[0])
			if err != nil {
				return nil, err
			}
			lo--
			hi, err := strconv.Atoi(r[1])
			if err != nil {
				return nil, err
			}
			hi--
			for i := lo; i <= hi; i++ {
				fields = append(fields, i)
			}
		} else {
			i, err := strconv.Atoi(f)
			if err != nil {
				return nil, err
			}
			i--
			fields = append(fields, i)
		}
	}
	return fields, nil
}

func main() {
	var (
		delimiter    string
		fieldsString string
		whitespace   bool
		suppress     bool
		parallel     int
	)

	app := cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "fields",
				Aliases:     []string{"f"},
				Required:    true,
				Value:       "",
				Usage:       "The list specifies fields, separated in the input by the field delimiter character (see the -d option). Output fields are separated by a single occurrence of the field delimiter character.",
				Destination: &fieldsString,
			},
			&cli.StringFlag{
				Name:        "delimiter",
				Aliases:     []string{"d"},
				Value:       "\t",
				Usage:       "Use delimiter as the field delimiter character instead of the tab character.",
				Destination: &delimiter,
			},
			&cli.BoolFlag{
				Name:        "whitespace",
				Aliases:     []string{"w"},
				Value:       false,
				Usage:       "Set the field delimiter to whitespace, which overrides the delimiter option. Output fields are separated by a single space character.",
				Destination: &whitespace,
			},
			&cli.BoolFlag{
				Name:        "suppress",
				Aliases:     []string{"s"},
				Value:       false,
				Usage:       "Suppress lines with no field delimiter characters. Unless specified, lines with no delimiters are passed through unmodified.",
				Destination: &suppress,
			},
			&cli.IntFlag{
				Name:        "parallel",
				Aliases:     []string{"p"},
				Value:       0,
				Usage:       "Specify the number of threads to use to process the input in parallel. The order of lines in the output will be different to the input.",
				Destination: &parallel,
			},
		},
		Action: func(c *cli.Context) error {
			file := c.Args().Get(0)
			switch {
			case file != "":
				in, err := os.Open(file)
				if err != nil {
					return err
				}
				defer in.Close()
				reader = bufio.NewScanner(in)
			default:
				reader = bufio.NewScanner(os.Stdin)
			}

			writer = bufio.NewWriter(os.Stdout)
			defer writer.Flush()

			if whitespace {
				delimiter = " "
			}

			cutter, err := NewCutter(delimiter, fieldsString, whitespace, suppress, parallel)
			if err != nil {
				return err
			}
			return cutter.cut()
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
