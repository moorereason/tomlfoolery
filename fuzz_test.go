//xxx go:build gofuzzbeta

package tomlfoolery

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"testing"

	tomltest "github.com/BurntSushi/toml-test"
)

func FuzzUnmarshal(f *testing.F) {
	addTomlTestCases := true

	if addTomlTestCases {
		skipTests := []string{
			"invalid/table/injection-1.toml",
			"invalid/table/injection-2.toml",
		}

		// Add embedded tests
		embFS := tomltest.EmbeddedTests()
		var i int
		fs.WalkDir(embFS, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				log.Fatal(err)
			}

			if !strings.HasSuffix(path, ".toml") {
				return nil
			}

			for _, s := range skipTests {
				if path == s {
					return nil
				}
			}

			b, err := fs.ReadFile(embFS, path)
			if err != nil {
				return err
			}

			fmt.Printf("add #%d: %s\n", i, path)
			f.Add(b)
			i++
			return nil
		})
	}

	for _, s := range []string{
		`[dog."tater.man"]
		type.name = "pug"`,
		`[ j . "ʞ" . 'l' ]`,
		`[[a.b]]
		a='b'`,
		"[table]\nhello = 'world'",
		`a=1979-05-27T00:32:00-07:00`,
		`a={f="1",b.c=3}`,
		`a="\\\n\t\""`,
		`a. b="c"`,
		`'"a"' = 1`,
		`"\"b\"" = 2`,
		`A = """\
							Test"""`,
		`a=1z=2`,
		`a="Name\tJos\u00E9\nLoc\tSF."`,
		`contributors = [
		  "Foo Bar <foo@example.com>",
		  { name = "Baz Qux", email = "bazqux@example.com", url = "https://example.com/bazqux" }
		]`,
		`foo = 2021-04-08`,
		`a=20x1-05-21`,
		`a=1_`,
		`a=0bfa`,
		`a=0o62`,
		`a=true`,
		`a=false`,
		`a=1__2`,
		`a=4e+9`,
		`a=-4e-2`,
		`a=inf`,
		`a=+inf`,
		`a=-inf`,
		`a=nan`,
		`a=+nan`,
		`a=-nan`,
		`a=[1, 2, "b"]`,
	} {
		f.Add([]byte(s))
	}

	aCmd, aCmdr := makeCommandParser(f, "A")
	bCmd, bCmdr := makeCommandParser(f, "B")

	f.Fuzz(func(t *testing.T, data []byte) {
		// FIXME: BurntSushi has some bugs.  Remove these checks later:
		//
		//   "\x00" - https://github.com/BurntSushi/toml/issues/317
		//   "\xff" - https://github.com/BurntSushi/toml/issues/317
		//   "\r" - https://github.com/BurntSushi/toml/issues/321
		if strings.ContainsAny(string(data), "\x00\xff\r") {
			t.Skip()
		}

		a, aOutErr, aErr := aCmdr.Encode(string(data))
		b, bOutErr, bErr := bCmdr.Encode(string(data))

		if aErr != nil {
			t.Fatal(aErr)
		}
		if bErr != nil {
			t.Fatal(bErr)
		}

		// If both decoders return an error, consider the input uninteresting.
		if aOutErr && bOutErr {
			return
		}

		if aOutErr != bOutErr {
			t.Fatalf("output errors differ:\ninput:\n\t%q\n%s (outErr=%t):\n%v\n%s (outErr=%t):\n%v\n", data, aCmd, aOutErr, a, bCmd, bOutErr, b)
		}

		// convert JSON to maps
		var amap map[string]interface{}
		err := json.Unmarshal([]byte(a), &amap)
		if err != nil {
			t.Fatal(err)
		}

		var bmap map[string]interface{}
		err = json.Unmarshal([]byte(b), &bmap)
		if err != nil {
			t.Fatal(err)
		}

		r := tomltest.Test{}

		rt := r.CompareJSON(amap, bmap)
		if rt.Failure != "" {
			t.Errorf("for input %q, output differs:\n%s:\n%v\n%s:\n%v\n\nFailure: %s\n", data, aCmd, a, bCmd, b, rt.Failure)
		}
	})
}

func makeCommandParser(f *testing.F, s string) (string, tomltest.CommandParser) {
	f.Helper()

	if len(s) != 1 {
		f.Fatalf("invalid use use makeCommandParser: length of s is %d", len(s))
	}

	ev := "TOML_" + s
	p := os.Getenv(ev)
	if len(p) == 0 {
		f.Fatalf("empty or missing environment variable: %s", ev)
	}

	_, err := os.Stat(p)
	if err != nil {
		f.Fatal(err)
	}

	return p, tomltest.NewCommandParser(nil, []string{p})
}
