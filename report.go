// Copyright (c) 2012 The Gocov Authors.
// 
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
// 
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
// 
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

package gocov

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"text/tabwriter"
)

type Report struct {
	packages []*Package
}

type reportFunction struct {
	*Function
	statementsReached int
}

type reportFunctionList []reportFunction

func (l reportFunctionList) Len() int {
	return len(l)
}

// TODO make sort method configurable?
func (l reportFunctionList) Less(i, j int) bool {
	var left, right float64
	if len(l[i].Statements) > 0 {
		left = float64(l[i].statementsReached) / float64(len(l[i].Statements))
	}
	if len(l[j].Statements) > 0 {
		right = float64(l[j].statementsReached) / float64(len(l[j].Statements))
	}
	if left < right {
		return true
	}
	return left == right && len(l[i].Statements) < len(l[j].Statements)
}

func (l reportFunctionList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

type reverse struct {
	sort.Interface
}

func (r reverse) Less(i, j int) bool {
	return r.Interface.Less(j, i)
}

// NewReport creates a new report.
func NewReport() (r *Report) {
	r = &Report{}
	return
}

// AddPackage adds a package's coverage information to the report.
func (r *Report) AddPackage(p *Package) {
	i := sort.Search(len(r.packages), func(i int) bool {
		return r.packages[i].Name >= r.packages[i].Name
	})
	if i < len(r.packages) && r.packages[i].Name == p.Name {
		panic("package already exists: result merging not implemented yet")
	} else {
		head := r.packages[:i]
		tail := append([]*Package{p}, r.packages[i:]...)
		r.packages = append(head, tail...)
	}
}

// Clear clears the coverage information from the report.
func (r *Report) Clear() {
	r.packages = nil
}

// PrintReport prints a coverage report to the given writer.
func PrintReport(w io.Writer, r *Report) {
	w = tabwriter.NewWriter(w, 0, 8, 0, '\t', 0)
	//fmt.Fprintln(w, "Package\tFunction\tStatements\t")
	//fmt.Fprintln(w, "-------\t--------\t---------\t")
	for _, pkg := range r.packages {
		printPackage(w, pkg)
		fmt.Fprintln(w)
	}
}

func printPackage(w io.Writer, pkg *Package) {
	functions := make(reportFunctionList, len(pkg.Functions))
	for i, fn := range pkg.Functions {
		reached := 0
		for _, stmt := range fn.Statements {
			if stmt.Reached > 0 {
				reached++
			}
		}
		functions[i] = reportFunction{fn, reached}
	}
	sort.Sort(reverse{functions})

	for _, fn := range functions {
		reached := fn.statementsReached
		var stmtPercent float64 = 0
		if len(fn.Statements) > 0 {
			stmtPercent = float64(reached) / float64(len(fn.Statements)) * 100
		}
		fmt.Fprintf(w, "%s/%s:%d\t %s\t %.2f%% (%d/%d)\n",
			pkg.Name, filepath.Base(fn.File), fn.Line, fn.Name, stmtPercent,
			reached, len(fn.Statements))
	}
}
