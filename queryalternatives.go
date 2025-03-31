package queryalternatives

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
)

// Alternative represents an alternative for a specific command.
type Alternative struct {
	// Path is the path to the alternative.
	Path string
	// Priority is the priority of the alternative.
	// Higher numbers indicate higher priority.
	Priority int
	// Slaves is a map of slave links to their corresponding paths.
	// Slaves are additional files that are linked to this alternative.
	Slaves map[string]string
}

// Alternatives represents the output of the `update-alternatives --query` command.
// It contains information about the alternatives for a specific command.
type Alternatives struct {
	// Name is the name of the alternatives group.
	// For example, "java" for the Java alternatives.
	Name string
	// Link is the generic path to the alternative.
	// For example, "/usr/bin/java" for the Java alternatives.
	Link string
	// Slaves is a map of slave links to their corresponding paths.
	Slaves map[string]string
	// Status indicates the status of the alternatives group.
	// It can be "auto" or "manual".
	// "auto" means the system will automatically select the best alternative.
	// "manual" means the user has manually selected an alternative.
	Status string
	// Best is the best alternative selected by the system.
	// It is the path to the best alternative.
	Best string
	// Value is the path to the alternative which is currently selected.
	// "none" means no alternative is selected.
	Value string
	// Alternatives is alternatives for this group.
	Alternatives []Alternative
}

type ParseError struct {
	Message string
	Line    int
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("error parsing alternatives: %d: %s", err.Line, err.Message)
}

func newAlternative() *Alternative {
	return &Alternative{
		Slaves: make(map[string]string),
	}
}

func newAlternatives() *Alternatives {
	return &Alternatives{
		Slaves:       make(map[string]string),
		Alternatives: make([]Alternative, 0),
	}
}

type Parser struct {
	R      *bufio.Reader
	lineNo int
}

func NewParser(r io.Reader) *Parser {
	if br, ok := r.(*bufio.Reader); ok {
		return &Parser{
			R:      br,
			lineNo: 0,
		}
	}
	return &Parser{
		R:      bufio.NewReader(r),
		lineNo: 0,
	}
}

func (r *Parser) readKeyValue() (string, string, error) {
	var line []byte
	var err error
	for {
		line, err = r.R.ReadBytes('\n')
		line = bytes.TrimRight(line, "\r\n")
		if err != nil {
			if err == io.EOF {
				if len(line) == 0 {
					return "", "", err
				}
			} else {
				return "", "", err
			}
		}
		r.lineNo++

		if len(line) != 0 {
			break
		}
	}

	parts := bytes.SplitN(line, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", &ParseError{
			Message: "malformed line",
			Line:    r.lineNo,
		}
	}

	key := string(parts[0])
	var value strings.Builder
	value.Write(bytes.TrimRight(bytes.TrimLeft(parts[1], " "), "\r\n"))

	for {
		next, err := r.R.Peek(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", "", err
		}
		if next[0] != ' ' {
			break
		}

		line, err = r.R.ReadBytes('\n')
		line = bytes.TrimRight(bytes.TrimLeft(line, " "), "\r\n")
		if err != nil {
			if err == io.EOF {
				if len(line) == 0 {
					break
				}
			} else {
				return "", "", err
			}
		}
		r.lineNo++

		if value.Len() > 0 {
			value.WriteByte('\n')
		}
		value.Write(line)
	}

	return key, value.String(), nil
}

func (r *Parser) parseSlaves(input string) (map[string]string, error) {
	slaves := make(map[string]string)
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			return nil, &ParseError{
				Message: "malformed slaves line",
				Line:    r.lineNo,
			}
		}
		slaves[parts[0]] = parts[1]
	}
	return slaves, nil
}

func (r *Parser) Parse() (*Alternatives, error) {
	result := newAlternatives()
	var currentAlt *Alternative

	for {
		k, v, err := r.readKeyValue()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if currentAlt == nil {
			switch k {
			case "Name":
				result.Name = v
			case "Link":
				result.Link = v
			case "Slaves":
				var err error
				result.Slaves, err = r.parseSlaves(v)
				if err != nil {
					return nil, err
				}
			case "Status":
				result.Status = v
			case "Best":
				result.Best = v
			case "Value":
				result.Value = v
			case "Alternative":
				currentAlt = newAlternative()
				currentAlt.Path = v
			default:
				return nil, &ParseError{
					Message: fmt.Sprintf("unexpected key: %s", k),
					Line:    r.lineNo,
				}
			}
		} else {
			switch k {
			case "Priority":
				priority, err := strconv.Atoi(v)
				if err != nil {
					return nil, &ParseError{
						Message: "invalid priority value",
						Line:    r.lineNo,
					}
				}
				currentAlt.Priority = priority
			case "Slaves":
				var err error
				currentAlt.Slaves, err = r.parseSlaves(v)
				if err != nil {
					return nil, err
				}
			case "Alternative":
				// Save the previous alternative before starting a new one
				result.Alternatives = append(result.Alternatives, *currentAlt)

				currentAlt = newAlternative()
				currentAlt.Path = v
			default:
				return nil, &ParseError{
					Message: fmt.Sprintf("unexpected key: %s", k),
					Line:    r.lineNo,
				}
			}
		}
	}

	if currentAlt != nil {
		// Save the last alternative
		result.Alternatives = append(result.Alternatives, *currentAlt)
	}

	return result, nil
}

// ParseString parses a string and returns an Alternatives object.
func ParseString(input string) (*Alternatives, error) {
	return NewParser(strings.NewReader(input)).Parse()
}

type QueryError struct {
	ExitStatus int
	Message    string
}

func (e *QueryError) Error() string {
	return "error querying alternatives: " + e.Message
}

// Query executes the `update-alternatives --query` command and returns the parsed result.
func Query(ctx context.Context, query string) (*Alternatives, error) {
	cmd := exec.CommandContext(ctx, "update-alternatives", "--query", query)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	defer stdout.Close()

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	result, err := NewParser(stdout).Parse()

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, &QueryError{
				ExitStatus: exitErr.ExitCode(),
				Message:    string(exitErr.Stderr),
			}
		}
		return nil, err
	}

	return result, err
}
