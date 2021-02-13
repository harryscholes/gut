package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
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
	parallel   int
}

func NewCutter(
	delimiter string,
	fields string,
	whitespace bool,
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
	var tokens []string

	for reader.Scan() {
		line := reader.Text()
		if c.whitespace {
			tokens = strings.Fields(line)
		} else {
			tokens = strings.Split(line, c.delimiter)
		}
		if len(tokens) < len(c.fields) {
			return errors.New("Ooops")
		}

		for _, field := range c.fields[:len(c.fields)-1] {
			fmt.Fprint(writer, tokens[field], c.delimiter)
		}
		fmt.Fprint(writer, tokens[c.fields[len(c.fields)-1]], "\n")
	}

	if err := reader.Err(); err != nil {
		return err
	}
	return nil
}

func (c *params) extractParallel() error {
	in := make(chan string)
	out := make(chan string)
	semaphore := make(chan struct{}, c.parallel)
	done := make(chan struct{})
	var wg sync.WaitGroup

	go func() {
		for reader.Scan() {
			in <- reader.Text()
		}
		close(in)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for line := range in {
			wg.Add(1)
			go func(line string) {
				defer wg.Done()
				semaphore <- struct{}{}

				var tokens []string
				if c.whitespace {
					tokens = strings.Fields(line)
				} else {
					tokens = strings.Split(line, c.delimiter)
				}

				// TODO add error handling
				// if len(tokens) < len(c.fields) {
				// 	return errors.New("Ooops")
				// }

				subset := make([]string, len(c.fields))
				for i, f := range c.fields {
					subset[i] = tokens[f]
				}
				joined := strings.Join(subset, c.delimiter)

				out <- joined
				<-semaphore
			}(line)
		}
	}()

	go func() {
		for joined := range out {
			fmt.Fprint(writer, joined, "\n")
		}
		done <- struct{}{}
	}()

	wg.Wait()
	close(out)
	<-done

	if err := reader.Err(); err != nil {
		return err
	}
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

			cutter, err := NewCutter(delimiter, fieldsString, whitespace, parallel)
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
