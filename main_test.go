// This file is part of licensed-notice-deduplicate
//
// SPDX-FileCopyrightText: Arduino s.r.l. and/or its affiliated companies
// SPDX-License-Identifier: GPL-3.0-or-later

package main

import (
	"os"
	"strings"
	"testing"
)

func deduplicate(t *testing.T, notice string) string {
	t.Helper()
	header, entries, err := parseNotice(strings.NewReader(notice))
	if err != nil {
		t.Fatalf("parseNotice: %v", err)
	}
	var out strings.Builder
	writeNotice(&out, header, groupByLicense(entries))
	return out.String()
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestGolden(t *testing.T) {
	got := deduplicate(t, readFile(t, "testdata/NOTICE"))
	if want := readFile(t, "testdata/NOTICE.golden"); got != want {
		t.Errorf("output differs from testdata/NOTICE.golden:\n%s", got)
	}
}

func TestIdempotent(t *testing.T) {
	once := readFile(t, "testdata/NOTICE.golden")
	if twice := deduplicate(t, once); twice != once {
		t.Errorf("a second run changed the output:\n%s", twice)
	}
}

func TestNoDuplicatesUnchanged(t *testing.T) {
	notice := "THIRD PARTY NOTICES\n\n*****\na@1.0.0\n\nMIT text\n\n*****\nb@2.0.0\n\nISC text\n"
	if got := deduplicate(t, notice); got != notice {
		t.Errorf("a notice without duplicates changed:\n%s", got)
	}
}

func TestTrailingSeparator(t *testing.T) {
	got := deduplicate(t, "THIRD PARTY NOTICES\n\n*****\na@1.0.0\n\nMIT text\n*****")
	if want := "THIRD PARTY NOTICES\n\n*****\na@1.0.0\n\nMIT text\n"; got != want {
		t.Errorf("got:\n%q\nwant:\n%q", got, want)
	}
}
