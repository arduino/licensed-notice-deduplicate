// This file is part of licensed-notice-deduplicate
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

// licensed-notice-deduplicate reads a NOTICE file produced by `licensed notice` and
// replaces it with a deduplicated version in place.
//
// All entries that share exactly the same license text are merged into a single
// block that lists every covered package, regardless of whether they belong to
// the same repository or not.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"
)

const groupedHeaderString = "License for the following packages:"

type entry struct {
	name    string
	version string
	body    string // license text, leading/trailing blank lines stripped
}

func parseNotice(r io.Reader) (header string, entries []entry, err error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		return
	}

	// Everything before the first ***** is the header.
	firstSep := slices.Index(lines, "*****")
	if firstSep == -1 {
		header = strings.Join(lines, "\n")
		return
	}

	headerLines := lines[:firstSep]
	for len(headerLines) > 0 && strings.TrimSpace(headerLines[len(headerLines)-1]) == "" {
		headerLines = headerLines[:len(headerLines)-1]
	}
	header = strings.Join(headerLines, "\n")

	// Each ***** line starts a new entry.
	i := firstSep
	for i < len(lines) {
		if lines[i] != "*****" {
			i++
			continue
		}
		i++ // consume *****

		// Detect if the notice file has already been deduplicated by looking for the grouped header string.
		if strings.TrimSpace(lines[i]) == groupedHeaderString {
			return "", nil, errors.New("the notice file appears to have already been deduplicated")
		}

		// Skip blank lines before the name@version identifier.
		for i < len(lines) && strings.TrimSpace(lines[i]) == "" {
			i++
		}
		if i >= len(lines) {
			break
		}

		nameVersion := strings.TrimSpace(lines[i])
		i++

		// Collect body until next ***** or EOF.
		var bodyLines []string
		for i < len(lines) && lines[i] != "*****" {
			bodyLines = append(bodyLines, lines[i])
			i++
		}
		// Strip surrounding blank lines; they are structural formatting.
		for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[0]) == "" {
			bodyLines = bodyLines[1:]
		}
		for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[len(bodyLines)-1]) == "" {
			bodyLines = bodyLines[:len(bodyLines)-1]
		}

		name, version, _ := strings.Cut(nameVersion, "@")
		entries = append(entries, entry{name: name, version: version, body: strings.Join(bodyLines, "\n")})
	}

	return
}

type licenseGroup struct {
	body    string
	members []entry
}

func groupByLicense(entries []entry) []licenseGroup {
	seen := make(map[string]int) // body -> index in groups
	var groups []licenseGroup

	for _, e := range entries {
		if idx, ok := seen[e.body]; ok {
			groups[idx].members = append(groups[idx].members, e)
		} else {
			seen[e.body] = len(groups)
			groups = append(groups, licenseGroup{body: e.body, members: []entry{e}})
		}
	}
	return groups
}

func writeNotice(w io.Writer, header string, groups []licenseGroup) {
	bw := bufio.NewWriter(w)
	defer bw.Flush()

	fmt.Fprintln(bw, header)

	for _, g := range groups {
		fmt.Fprintln(bw)
		fmt.Fprintln(bw, "*****")

		if len(g.members) == 1 {
			e := g.members[0]
			id := e.name
			if e.version != "" {
				id += "@" + e.version
			}
			fmt.Fprintln(bw, id)
		} else {
			fmt.Fprintln(bw, groupedHeaderString)
			for _, e := range g.members {
				id := e.name
				if e.version != "" {
					id += "@" + e.version
				}
				fmt.Fprintf(bw, "  %s\n", id)
			}
		}

		fmt.Fprintln(bw)
		fmt.Fprintln(bw, g.body)
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: licensed-notice-deduplicate .licenses/NOTICE.project\n")
		os.Exit(1)
	}

	// Parse the notice file
	in, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening notice file: %v\n", err)
		os.Exit(1)
	}
	header, entries, err := parseNotice(in)
	in.Close()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing notice: %v\n", err)
		os.Exit(1)
	}

	// Group entries by identical license text.
	groups := groupByLicense(entries)

	// Write the deduplicated notice back to the same file.
	out, err := os.Create(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening notice file for writing: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()
	writeNotice(out, header, groups)
}
