//xxx go:build gofuzzbeta

package tomlfoolery

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"testing"

	tomltest "github.com/BurntSushi/toml-test"
)

func FuzzUnmarshal(f *testing.F) {
	addTomlTestCases := false

	if addTomlTestCases {
		skipTests := []string{
			"valid/datetime/datetime.toml",
			"valid/datetime/local.toml",
			"valid/datetime/local-time.toml",
			"valid/float/zero.toml",
			"valid/string/multiline.toml",
			"valid/string/multiline-quotes.toml",
			"invalid/datetime/impossible-date.toml",
			"invalid/float/exp-leading-us.toml",
			"invalid/float/exp-point-2.toml",
			"invalid/float/leading-point-neg.toml",
			"invalid/float/leading-point-plus.toml",
			"invalid/float/leading-zero-neg.toml",
			"invalid/float/leading-zero-plus.toml",
			"invalid/float/us-after-point.toml",
			"invalid/float/us-before-point.toml",
			"invalid/integer/leading-zero-sign-1.toml",
			"invalid/integer/leading-zero-sign-2.toml",
			"invalid/string/multiline-escape-space.toml",
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
		//`a=1979-05-27T00:32:00-07:00`,
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
		// `foo = 2021-04-08`,
		// `a=20x1-05-21`,
		`a=1_`,
		`a=0bfa`,
		`a=0o62`,
		`a=true`,
		`a=false`,
		`a=1__2`,
		`a=4e+9`,
		`a=-4e-2`,
		// `a=inf`,
		// `a=+inf`,
		// `a=-inf`,
		// `a=nan`,
		// `a=+nan`,
		// `a=-nan`,
		`a=[1, 2, "b"]`,
	} {
		f.Add([]byte(s))
	}

	aCmd, aCmdr := makeCommandParser(f, "A")
	bCmd, bCmdr := makeCommandParser(f, "B")

	f.Fuzz(func(t *testing.T, data []byte) {
		a, aOutErr, aErr := aCmdr.Encode(string(data))
		b, bOutErr, bErr := bCmdr.Encode(string(data))

		if aErr != nil {
			t.Fatal(aErr)
		}
		if bErr != nil {
			t.Fatal(bErr)
		}

		if aOutErr != bOutErr && !strings.ContainsRune(string(data), '\r') {
			t.Fatalf("output errors differ:\ninput:\n\t%q\n%s (outErr=%t):\n%v\n%s (outErr=%t):\n%v\n", data, aCmd, aOutErr, a, bCmd, bOutErr, b)
		}
		if aOutErr {
			return
		}

		r := tomltest.Test{}

		rt := r.CompareTOML(a, b)
		if rt.Failure != "" {
			fmt.Printf("for input %q, output differs:\n%s:\n%v\n%s:\n%v\n", data, a, aCmd, b, bCmd)
			t.Fatalf("%#v\n", rt)
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
