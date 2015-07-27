package dsfile

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var testCodecs = []Codec{
	new(Cbor),
	new(Gob),
	new(Json),
}

func TestBasics(t *testing.T) {
	for _, c := range testCodecs {
		testBasics(c, t)
	}
}

func testBasics(c Codec, t *testing.T) {
	type Object struct {
		Str   string
		Int   int64
		Slice []int
	}
	obj := Object{
		Str:   "foobar",
		Int:   42,
		Slice: []int{5, 3, 2, 1, 4},
	}
	path := filepath.Join(os.TempDir(), "dsfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	port := rand.Intn(20000) + 30000
	file, err := New(&obj, path, c, NewPortLocker(port))
	if err != nil {
		t.Fatalf("new %v", err)
	}
	err = file.Save()
	if err != nil {
		t.Fatalf("save %v", err)
	}
	file.Close()

	var obj2 Object
	file, err = New(&obj2, path, c, NewPortLocker(port))
	if err != nil {
		t.Fatalf("new %v", err)
	}
	defer file.Close()
	if obj2.Str != obj.Str {
		t.Fatalf("str not match")
	}
	if obj2.Int != obj.Int {
		t.Fatalf("int not match")
	}
	if len(obj2.Slice) != len(obj.Slice) {
		t.Fatalf("slice not match")
	}
	for i, n := range obj2.Slice {
		if n != obj.Slice[i] {
			t.Fatalf("slice not match")
		}
	}
}

func TestInvalidObject(t *testing.T) {
	for _, c := range testCodecs {
		testInvalidObject(c, t)
	}
}

func testInvalidObject(c Codec, t *testing.T) {
	_, err := New(42, "foo", c, NewPortLocker(0))
	if err == nil || err.Error() != "dsfile: object must be a pointer" {
		t.Fatalf("should fail, got %v", err)
	}
}

func TestLockFail(t *testing.T) {
	for _, c := range testCodecs {
		testLockFail(c, t)
	}
}

func testLockFail(c Codec, t *testing.T) {
	port := rand.Intn(20000) + 30000
	path := filepath.Join(os.TempDir(), "dsfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	obj := struct{}{}
	f1, err := New(&obj, path, c, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}
	_, err = New(&obj, path, c, NewPortLocker(port))
	if err == nil || err.Error() != "dsfile: lock fail" {
		t.Fatal("should fail")
	}
	f1.Close()
}

func TestLockFail2(t *testing.T) {
	for _, c := range testCodecs {
		testLockFail2(c, t)
	}
}

func testLockFail2(c Codec, t *testing.T) {
	lockFilePath := filepath.Join(os.TempDir(), fmt.Sprintf("dsfile-test-lock-%d", rand.Int63()))
	path := filepath.Join(os.TempDir(), "dsfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	obj := struct{}{}
	f1, err := New(&obj, path, c, NewFileLocker(lockFilePath))
	if err != nil {
		t.Fatal(err)
	}
	defer f1.Close()
	_, err = New(&obj, path, c, NewFileLocker(lockFilePath))
	if err == nil || err.Error() != "dsfile: lock fail" {
		t.Fatal("should fail")
	}
}

func TestLockFail3(t *testing.T) {
	for _, c := range testCodecs {
		testLockFail3(c, t)
	}
}

func testLockFail3(c Codec, t *testing.T) {
	lockFilePath := os.TempDir()
	path := filepath.Join(os.TempDir(), "dsfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	obj := struct{}{}
	func() {
		defer func() {
			if err := recover(); err == nil || !strings.HasPrefix(err.(string), fmt.Sprintf("open lock file %s error", os.TempDir())) {
				fmt.Printf("%v\n", err.(string))
				t.Fatal("should fail")
			}
		}()
		New(&obj, path, c, NewFileLocker(lockFilePath))
	}()
}

func TestCorruptedFile(t *testing.T) {
	for _, c := range testCodecs {
		testCorruptedFile(c, t)
	}
}

func testCorruptedFile(c Codec, t *testing.T) {
	obj := map[string]string{
		"1": "foo",
		"2": "bar",
		"3": "baz",
	}
	port := rand.Intn(20000) + 30000
	path := filepath.Join(os.TempDir(), "dsfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	file, err := New(&obj, path, c, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}
	err = file.Save()
	if err != nil {
		t.Fatal(err)
	}
	file.Close()

	content, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = content[:len(content)/2]
	err = ioutil.WriteFile(path, content, 0644)
	if err != nil {
		t.Fatal(err)
	}
	file, err = New(&obj, path, c, NewPortLocker(port))
	if err == nil || !strings.HasPrefix(err.Error(), "dsfile: decode error") {
		t.Fatalf("should fail, got %v", err)
	}
}

func TestSaveFail(t *testing.T) {
	for _, c := range testCodecs {
		testSaveFail(c, t)
	}
}

func testSaveFail(c Codec, t *testing.T) {
	dir := filepath.Join(os.TempDir(), fmt.Sprintf("dsfile-testdir-%d", rand.Int63()))
	os.Mkdir(dir, 0755)
	path := filepath.Join(dir, "foo")
	port := rand.Intn(20000) + 30000
	obj := struct{}{}
	file, err := New(&obj, path, c, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}
	os.Chmod(dir, 0000)
	err = file.Save()
	if err == nil || !strings.HasPrefix(err.Error(), "dsfile: open temp file") {
		t.Fatal("should fail")
	}
}

func TestSaveFail2(t *testing.T) {
	for _, c := range testCodecs {
		testSaveFail2(c, t)
	}
}

func testSaveFail2(c Codec, t *testing.T) {
	port := rand.Intn(20000) + 30000
	path := filepath.Join(os.TempDir(), "dsfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	obj := struct{}{}
	file, err := New(&obj, path, c, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(path)
	os.Mkdir(path, 0000)
	err = file.Save()
	if err == nil || !strings.HasPrefix(err.Error(), "dsfile: rename temp file") {
		t.Fatal("should fail")
	}
}

func TestEncodeError(t *testing.T) {
	for _, c := range []Codec{
		new(Gob),
		new(Json),
	} {
		testEncodeError(c, t)
	}
}

func testEncodeError(c Codec, t *testing.T) {
	port := rand.Intn(20000) + 30000
	path := filepath.Join(os.TempDir(), "gobfile-test-"+strconv.FormatInt(rand.Int63(), 10))
	obj := make(chan bool)
	file, err := New(&obj, path, c, NewPortLocker(port))
	if err != nil {
		t.Fatal(err)
	}
	err = file.Save()
	if err == nil || !strings.HasPrefix(err.Error(), "dsfile: encode error") {
		t.Fatal("should fail")
	}
}
