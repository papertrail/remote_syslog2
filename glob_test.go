package main

import (
	"io/ioutil"
	"os"
	"regexp"

	. "gopkg.in/check.v1"
)

type GlobSuite struct {
	wr      *WorkerRegistry
	tempdir string
}

var _ = Suite(&GlobSuite{})

func (s *GlobSuite) SetUpSuite(c *C) {
	s.wr = NewWorkerRegistry()
}

func (s *GlobSuite) TearDownSuite(c *C) {
}

func (s *GlobSuite) SetUpTest(c *C) {
}

func (s *GlobSuite) TearDownTest(c *C) {
	os.RemoveAll(s.tempdir)
}

func (s *GlobSuite) TestNoFiles(c *C) {
	dir, err := ioutil.TempDir("", "TestNoFiles")
	c.Assert(err, IsNil)
	s.tempdir = dir
	pats := []string{dir + "/*"}
	m, err := glob(pats, []*regexp.Regexp{}, s.wr, false)
	c.Assert(err, IsNil)
	c.Assert(m, HasLen, 0)
}

func (s *GlobSuite) TestSimple(c *C) {
	dir, err := ioutil.TempDir("", "TestSimple")
	c.Assert(err, IsNil)
	s.tempdir = dir
	f1, err := ioutil.TempFile(s.tempdir, "")
	c.Assert(err, IsNil)
	f2, err := ioutil.TempFile(s.tempdir, "")
	c.Assert(err, IsNil)
	// exact match
	pats := []string{f1.Name(), f2.Name()}
	a, err := glob(pats, []*regexp.Regexp{}, s.wr, false)
	c.Assert(err, IsNil)
	m := s.mkmap(a)
	c.Assert(m, HasLen, 2)
	c.Assert(m[f1.Name()], Equals, true)
	c.Assert(m[f2.Name()], Equals, true)
	// simple pattern
	pats = []string{dir + "/*"}
	a, err = glob(pats, []*regexp.Regexp{}, s.wr, false)
	c.Assert(err, IsNil)
	m = s.mkmap(a)
	c.Assert(m, HasLen, 2)
	c.Assert(m[f1.Name()], Equals, true)
	c.Assert(m[f2.Name()], Equals, true)
}

func (s *GlobSuite) TestNested(c *C) {
	dir, err := ioutil.TempDir("", "TestNested")
	c.Assert(err, IsNil)
	s.tempdir = dir
	f1, err := ioutil.TempFile(s.tempdir, "")
	c.Assert(err, IsNil)
	nested := dir + "/nested"
	c.Assert(os.Mkdir(nested, os.ModePerm), IsNil)
	f2, err := ioutil.TempFile(nested, "123XX123")
	c.Assert(err, IsNil)
	// top-level file, skip directory
	pats := []string{dir + "/*"}
	a, err := glob(pats, []*regexp.Regexp{}, s.wr, false)
	c.Assert(err, IsNil)
	m := s.mkmap(a)
	c.Assert(m, HasLen, 1)
	c.Assert(m[f1.Name()], Equals, true)
	// nested file
	pats = []string{dir + "/*/*123XX123*"}
	a, err = glob(pats, []*regexp.Regexp{}, s.wr, false)
	c.Assert(err, IsNil)
	m = s.mkmap(a)
	c.Assert(m, HasLen, 1)
	c.Assert(m[f2.Name()], Equals, true)
	// the whole 9 yards
	pats = []string{dir + "/**/*"}
	a, err = glob(pats, []*regexp.Regexp{}, s.wr, false)
	c.Assert(err, IsNil)
	m = s.mkmap(a)
	c.Assert(m, HasLen, 1)
	c.Assert(m[f2.Name()], Equals, true)
}

func (s *GlobSuite) mkmap(a []string) map[string]bool {
	m := map[string]bool{}
	for _, v := range a {
		m[v] = true
	}
	return m
}