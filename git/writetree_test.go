package git

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func hashString(str string) Sha1 {
	s, _, err := HashReader("blob", strings.NewReader(str))
	if err != nil {
		panic(err)
	}
	return s
}

func TestWriteIndex(t *testing.T) {
	testcases := []struct {
		IndexObjects []*IndexEntry
		Sha1         string
	}{
		{
			nil,
			// An empty tree hashes to this, not 0 (even with the official git client), because
			// of the type prefix in the blob.
			"4b825dc642cb6eb9a060e54bf8d69288fbee4904",
		},
		// Simple case, a single file
		{
			[]*IndexEntry{&IndexEntry{
				PathName: IndexPath("foo"),
				FixedIndexEntry: FixedIndexEntry{
					Mode:  ModeBlob,
					Fsize: 4,
					Sha1:  hashString("bar\n"),
				},
			},
			},
			"6a09c59ce8eb1b5b4f89450103e67ff9b3a3b1ae",
		},
		// Same as case 1, but with the executable bit set.
		{
			[]*IndexEntry{&IndexEntry{
				PathName: IndexPath("foo"),
				FixedIndexEntry: FixedIndexEntry{
					Mode:  ModeExec,
					Fsize: 4,
					Sha1:  hashString("bar\n"),
				},
			},
			},
			"e10d3585c7b4bec6b573e40d6a0c097a7e790abe",
		},
		// A symlink from bar to foo.
		{
			[]*IndexEntry{&IndexEntry{
				PathName: IndexPath("bar"),
				FixedIndexEntry: FixedIndexEntry{
					Mode:  ModeSymlink,
					Fsize: 3,
					Sha1:  hashString("foo"),
				},
			},
			},
			"985badfa7a966612b9f9adadbaa6a30aa3e0b1f5",
		},
		// Simple case, two files
		{
			[]*IndexEntry{
				&IndexEntry{
					PathName: IndexPath("bar"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("bar\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("foo"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("foo\n"),
					},
				},
			},
			"89ff1a2aefcbff0f09197f0fd8beeb19a7b6e51c",
		},
		// A single file in a subdirectory
		{
			[]*IndexEntry{&IndexEntry{
				PathName: IndexPath("foo/bar"),
				FixedIndexEntry: FixedIndexEntry{
					Mode:  ModeBlob,
					Fsize: 4,
					Sha1:  hashString("bar\n"),
				},
			},
			},
			"7b74f9ae4e4f7232e386fd8bcb9a240e6713fadf",
		},
		// Two files in a subdirectory
		{
			[]*IndexEntry{
				&IndexEntry{
					PathName: IndexPath("foo/bar"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("bar\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("foo/foo"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("foo\n"),
					},
				},
			},
			"e3331a4b901802f18658544c4ae320de93ab14ef",
		},
		// Both a file and a subtree
		{
			[]*IndexEntry{
				&IndexEntry{
					PathName: IndexPath("bar"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("bar\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("foo/foo"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("foo\n"),
					},
				},
			},
			"17278814743a70ed99aca0271ecdf5b544f10e5b",
		},
		// A file and a subtree with multiple entries
		{
			[]*IndexEntry{
				&IndexEntry{
					PathName: IndexPath("bar"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("bar\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("foo/bar"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("bar\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("foo/foo"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("foo\n"),
					},
				},
			},
			"18473c7faa0d4bb4913fd41a6768dbcf5fa70723",
		},
		// A deep subtree
		{
			[]*IndexEntry{
				&IndexEntry{
					PathName: IndexPath("foo/bar/baz"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("baz\n"),
					},
				},
			},
			"cc1846d0911b1790fd15859ffdf48598cb46b7b0",
		},
		// Two different subtrees
		{
			[]*IndexEntry{
				&IndexEntry{
					PathName: IndexPath("bar/bar"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("bar\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("foo/foo"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("foo\n"),
					},
				},
			},
			"65de833961e3dc313b13a2cf0a35a3bab772fc0b",
		},
		// Tree followed by a file.
		{
			[]*IndexEntry{
				&IndexEntry{
					PathName: IndexPath("bar/bar"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("bar\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("foo"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("foo\n"),
					},
				},
			},
			"615b1bd6b48087f25d16cc78279ea48ce5b1b59d",
		},
		{
			[]*IndexEntry{
				&IndexEntry{
					PathName: IndexPath("bar/bar"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("bar\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("baz/baz"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("baz\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("foo/foo"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("foo\n"),
					},
				},
			},
			"8b9f58ced67de613a7570726233ec83fa56a3d52",
		},
		// A file sandwiched between 2 trees
		{
			[]*IndexEntry{
				&IndexEntry{
					PathName: IndexPath("bar/bar"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("bar\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("baz"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("baz\n"),
					},
				},
				&IndexEntry{
					PathName: IndexPath("foo/foo"),
					FixedIndexEntry: FixedIndexEntry{
						Mode:  ModeBlob,
						Fsize: 4,
						Sha1:  hashString("foo\n"),
					},
				},
			},
			"18a6e5a95bb59e96dba722025de6abc692661bb6",
		},
	}

	gitdir, err := ioutil.TempDir("", "gitwriteindex")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(gitdir)

	c, err := NewClient(gitdir, "")
	if err != nil {
		t.Fatal(err)
	}

	for i, tc := range testcases {
		treeid, err := writeIndexEntries(c, "", tc.IndexObjects)
		if err != nil && err != ObjectExists {
			t.Error(err)
			continue
		}

		expected, err := Sha1FromString(tc.Sha1)
		if err != nil {
			t.Fatal(err)
		}
		if treeid != TreeID(expected) {
			t.Errorf("Unexpected hash for test case %d: got %v want %v", i, treeid, expected)
		}

	}
}
