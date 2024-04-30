package file_test

import (
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fleetdm/fleet/v4/pkg/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	// Setup
	originalPath := filepath.Join(tmp, "original")
	dstPath := filepath.Join(tmp, "copy")
	expectedContents := []byte("foo")
	expectedMode := fs.FileMode(0644)
	require.NoError(t, os.WriteFile(originalPath, expectedContents, os.ModePerm))
	require.NoError(t, os.WriteFile(dstPath, []byte("this should be overwritten"), expectedMode))

	// Test
	require.NoError(t, file.Copy(originalPath, dstPath, expectedMode))

	contents, err := os.ReadFile(originalPath)
	require.NoError(t, err)
	assert.Equal(t, expectedContents, contents)

	contents, err = os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, expectedContents, contents)

	info, err := os.Stat(dstPath)
	require.NoError(t, err)
	assert.Equal(t, expectedMode, info.Mode())

	// Copy of nonexistent path fails
	require.Error(t, file.Copy(filepath.Join(tmp, "notexist"), dstPath, os.ModePerm))

	// Copy to nonexistent directory
	require.Error(t, file.Copy(originalPath, filepath.Join("tmp", "notexist", "foo"), os.ModePerm))
}

func TestCopyWithPerms(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	// Setup
	originalPath := filepath.Join(tmp, "original")
	dstPath := filepath.Join(tmp, "copy")
	expectedContents := []byte("foo")
	expectedMode := fs.FileMode(0755)
	require.NoError(t, os.WriteFile(originalPath, expectedContents, expectedMode))

	// Test
	require.NoError(t, file.CopyWithPerms(originalPath, dstPath))

	contents, err := os.ReadFile(dstPath)
	require.NoError(t, err)
	assert.Equal(t, expectedContents, contents)

	info, err := os.Stat(dstPath)
	require.NoError(t, err)
	assert.Equal(t, expectedMode, info.Mode())
}

func TestExists(t *testing.T) {
	t.Parallel()
	tmp := t.TempDir()

	// Setup
	path := filepath.Join(tmp, "file")
	require.NoError(t, os.WriteFile(path, []byte(""), os.ModePerm))
	require.NoError(t, os.MkdirAll(filepath.Join(tmp, "dir", "nested"), os.ModePerm))

	// Test
	exists, err := file.Exists(path)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = file.Exists(filepath.Join(tmp, "notexist"))
	require.NoError(t, err)
	assert.False(t, exists)

	exists, err = file.Exists(filepath.Join(tmp, "dir"))
	require.NoError(t, err)
	assert.False(t, exists)
}

// TestExtractInstallerMetadata tests the ExtractInstallerMetadata function. It
// calls the function for every file under testdata/installers and checks that
// it returns the expected metadata by comparing it to the software name,
// version and hash in the filename.
//
// The filename should have the following format:
//
//	<software_name>$<version>$<sha256hash>[$<anything>].<extension>
//
// That is, it breaks the file name at the dollar sign and the first part is
// the expected name, the second is the expected version, the third is the
// hex-encoded hash. Note that by default, files in testdata/installers are NOT
// included in git, so the test files must be added manually (for size and
// licenses considerations). Why the dollar sign? Because dots, dashes and
// underlines are more likely to be part of the name or version.
func TestExtractInstallerMetadata(t *testing.T) {
	dents, err := os.ReadDir(filepath.Join("testdata", "installers"))
	if err != nil {
		t.Fatal(err)
	}

	for _, dent := range dents {
		if !dent.Type().IsRegular() || strings.HasPrefix(dent.Name(), ".") {
			continue
		}
		t.Run(dent.Name(), func(t *testing.T) {
			parts := strings.Split(strings.TrimSuffix(dent.Name(), filepath.Ext(dent.Name())), "$")
			if len(parts) < 3 {
				t.Fatalf("invalid filename, expected at least 3 sections, got %d: %s", len(parts), dent.Name())
			}
			wantName, wantVersion, wantHash := parts[0], parts[1], parts[2]

			f, err := os.Open(filepath.Join("testdata", "installers", dent.Name()))
			require.NoError(t, err)
			defer f.Close()

			name, version, hash, err := file.ExtractInstallerMetadata(dent.Name(), f)
			require.NoError(t, err)
			assert.Equal(t, wantName, name)
			assert.Equal(t, wantVersion, version)
			assert.Equal(t, wantHash, hex.EncodeToString(hash))
		})
	}
}
